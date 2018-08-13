package build

import (
	"time"

	"github.com/asdine/storm"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

// Event constants
const (
	EventBuildCreate = "build:new"
	EventBuildUpdate = "build:update"
	EventBuildDelete = "build:delete"
)

type BuildManager struct {
	db            *buildStormDB
	stepManager   *StepManager
	streamManager *StreamManager
	brokers       []domain.Broker
}

func NewBuildManager(
	db *storm.DB,
	stepManager *StepManager,
	streamManager *StreamManager,
) *BuildManager {
	m := &BuildManager{
		db:            newBuildStormDB(db),
		stepManager:   stepManager,
		streamManager: streamManager,
		brokers:       []domain.Broker{},
	}
	return m
}

func (m *BuildManager) AddBroker(b domain.Broker) {
	m.brokers = append(m.brokers, b)
}

func (m *BuildManager) Create(
	t *task.Task,
	params map[string]string,
) (*Build, *domain.ValidationErrors) {
	// TODO: implement validation
	timestamp := time.Now().UTC()
	b := &Build{
		ID:         uuid.NewV3(uuid.NewV1(), t.ID).String(),
		Task:       t,
		Parameters: params,
		CreatedAt:  timestamp,
		UpdatedAt:  timestamp,
		Status:     velocity.StateWaiting,
	}
	m.db.save(b)

	// steps := []*Step{}
	for i, tS := range t.VTask.Steps {
		step := m.stepManager.create(b, i, &tS)

		for _, streamName := range tS.GetOutputStreams() {
			m.streamManager.create(step, streamName)
			// step.Streams = append(step.Streams, stream)
		}
		// steps = append(steps, step)
	}
	// b.Steps = steps

	for _, br := range m.brokers {
		br.EmitAll(&domain.Emit{
			Event:   EventBuildCreate,
			Payload: b,
		})
	}

	return b, nil
}

func (m *BuildManager) Update(b *Build) error {
	if err := m.db.save(b); err != nil {
		return err
	}
	for _, br := range m.brokers {
		br.EmitAll(&domain.Emit{
			Event:   EventBuildUpdate,
			Payload: b,
		})
	}

	return nil
}

func (m *BuildManager) GetBuildByID(id string) (*Build, error) {
	return GetBuildByID(m.db.DB, id)
}

func (m *BuildManager) GetAllForProject(p *project.Project, q *BuildQuery) ([]*Build, int) {
	return m.db.getAllForProject(p, q)
}

func (m *BuildManager) GetAllForCommit(c *githistory.Commit, q *BuildQuery) ([]*Build, int) {
	return m.db.getAllForCommit(c, q)
}

// func (m *BuildManager) GetAllForBranch(b *githistory.Branch, q *domain.PagingQuery) ([]*Build, int) {
// 	return m.db.getAllForBranch(b, q)
// }

func (m *BuildManager) GetAllForTask(t *task.Task, q *BuildQuery) ([]*Build, int) {
	return m.db.getAllForTask(t, q)
}

func (m *BuildManager) GetRunningBuilds() ([]*Build, int) {
	return m.db.getRunningBuilds()
}

func (m *BuildManager) GetWaitingBuilds() ([]*Build, int) {
	return m.db.getWaitingBuilds()
}
