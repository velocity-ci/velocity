package slave

import (
	"fmt"
	"log"
	"os"

	"github.com/velocity-ci/velocity/backend/api/domain/build"
)

type Manager struct {
	logger       *log.Logger
	slaves       map[string]Slave
	buildManager build.Repository
}

func NewManager(buildManager build.Repository) *Manager {
	return &Manager{
		logger:       log.New(os.Stdout, "[manager:slave]", log.Lshortfile),
		slaves:       map[string]Slave{},
		buildManager: buildManager,
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

func (m *Manager) GetSlaves() map[string]Slave {
	return m.slaves
}

func (m *Manager) Save(s Slave) {
	m.logger.Printf("saving slave: %s", s.ID)
	m.slaves[s.ID] = s
	m.logger.Printf("saved slave: %s\n", s.ID)
}

func (m *Manager) GetSlaveByID(slaveID string) (Slave, error) {
	if m.Exists(slaveID) {
		return m.slaves[slaveID], nil
	}
	return Slave{}, fmt.Errorf("could not find slave %s", slaveID)
}

func (m *Manager) StartBuild(slave Slave, b build.Build) {
	// TODO: Sync known hosts
	slave.State = "busy"
	m.Save(slave)
	m.logger.Printf("set slave %s as busy", slave.ID)

	b.Status = "running"
	m.buildManager.SaveBuild(b)
	m.logger.Printf("set build %s as running", b.ID)

	buildSteps, count := m.buildManager.GetBuildStepsForBuild(b)
	if count < 1 {
		m.logger.Printf("creating build steps for %s", b.ID)
		for i, s := range b.Task.VTask.Steps {
			bS := build.NewBuildStep(
				b,
				uint64(i),
				s,
			)
			m.buildManager.SaveBuildStep(bS)
			for _, oS := range bS.Step.GetOutputStreams() {
				m.buildManager.SaveOutputStream(oS)
			}
			buildSteps = append(buildSteps, bS)
		}
	}
	slave.Command = NewBuildCommand(b, buildSteps)

	slave.ws.WriteJSON(slave.Command)
}
