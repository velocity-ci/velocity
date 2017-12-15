package slave

import (
	"fmt"
	"log"
	"os"

	"github.com/velocity-ci/velocity/backend/api/domain/knownhost"

	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"

	"github.com/velocity-ci/velocity/backend/api/domain/build"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
	"github.com/velocity-ci/velocity/backend/api/websocket"
)

type Manager struct {
	logger           *log.Logger
	slaves           map[string]Slave
	buildManager     build.Repository
	taskManager      task.Repository
	commitManager    commit.Repository
	projectManager   project.Repository
	knownHostManager knownhost.Repository
	websocketManager *websocket.Manager
}

func NewManager(
	buildManager build.Repository,
	taskManager task.Repository,
	commitManager commit.Repository,
	projectManager project.Repository,
	knownHostManager knownhost.Repository,
	websocketManager *websocket.Manager,
) *Manager {
	return &Manager{
		logger:           log.New(os.Stdout, "[manager:slave]", log.Lshortfile),
		slaves:           map[string]Slave{},
		buildManager:     buildManager,
		taskManager:      taskManager,
		commitManager:    commitManager,
		projectManager:   projectManager,
		websocketManager: websocketManager,
	}
}

func (m *Manager) Exists(slaveID string) bool {
	if _, ok := m.slaves[slaveID]; ok {
		return true
	}
	return false
}

func (m *Manager) WebSocketConnected(slaveID string) bool {
	if m.Exists(slaveID) {
		if m.slaves[slaveID].ws != nil {
			return true
		}
	}
	return false
}

func (m *Manager) GetSlaves(q SlaveQuery) ([]Slave, uint64) {
	slaves := []Slave{}
	skipCounter := uint64(0)
	for _, v := range m.slaves {
		if uint64(len(slaves)) >= q.Amount {
			break
		}
		if skipCounter < (q.Page-1)*q.Amount {
			skipCounter++
			break
		}
		if q.Status == "all" || q.Status == v.State {
			slaves = append(slaves, v)
		}
	}

	return slaves, uint64(len(slaves)) + skipCounter*q.Amount
}

func (m *Manager) Save(s Slave) {
	var ev string
	if m.Exists(s.ID) {
		ev = websocket.VUpdateSlave
	} else {
		ev = websocket.VNewSlave
	}
	m.slaves[s.ID] = s
	m.logger.Printf("saved slave: %s\n", s.ID)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   "slaves",
		Event:   ev,
		Payload: s,
	})
}

func (m *Manager) GetSlaveByID(slaveID string) (Slave, error) {
	if m.Exists(slaveID) {
		return m.slaves[slaveID], nil
	}
	return Slave{}, fmt.Errorf("could not find slave %s", slaveID)
}

func (m *Manager) StartBuild(slave Slave, b build.Build) {
	slave.State = "busy"
	m.Save(slave)
	m.logger.Printf("set slave %s as busy", slave.ID)

	knownHosts, _ := m.knownHostManager.GetAll(knownhost.KnownHostQuery{})

	slave.Command = NewKnownHostCommand(knownHosts)
	slave.ws.WriteJSON(slave.Command)
	m.logger.Printf("Sent to slave %s: %v", slave.ID, slave.Command)

	b.Status = "running"
	m.buildManager.UpdateBuild(b)
	m.logger.Printf("set build %s as running", b.ID)

	t, err := m.taskManager.GetByTaskID(b.TaskID)
	if err != nil {
		m.logger.Fatalf("task %s not found for build %s?!?!", b.TaskID, b.ID)
	}

	c, err := m.commitManager.GetCommitByCommitID(t.CommitID)
	if err != nil {
		m.logger.Fatalf("commit %s not found for task %s?!?!", t.CommitID, t.ID)
	}

	p, err := m.projectManager.GetByID(c.ProjectID)
	if err != nil {
		m.logger.Fatalf("project %s not found for commit %s?!?!", c.ProjectID, c.ID)
	}

	buildSteps, count := m.buildManager.GetBuildStepsByBuildID(b.ID)
	if count > 1 {
		// Remove existing buildSteps
		for _, bS := range buildSteps {
			m.buildManager.DeleteBuildStep(bS)
		}
		buildSteps = []build.BuildStep{}
	}
	for i, s := range t.Steps {
		bS := build.NewBuildStep(
			b.ID,
			uint64(i),
		)
		m.buildManager.CreateBuildStep(bS)
		buildSteps = append(buildSteps, bS)

		for _, streamName := range s.GetOutputStreams() {
			stream := build.NewBuildStepStream(bS.ID, streamName)
			m.buildManager.SaveStream(stream)
		}
		m.logger.Printf("created streams for %s", bS.ID)
	}
	m.logger.Printf("created build steps for %s", b.ID)
	slave.Command = NewBuildCommand(b, buildSteps, p, c, t)

	slave.ws.WriteJSON(slave.Command)

	m.logger.Printf("Sent to slave %s: %v", slave.ID, slave.Command)
}
