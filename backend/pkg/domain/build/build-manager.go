package build

import (
	"time"

	"github.com/asdine/storm"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type BuildManager struct {
	db            *buildStormDB
	stepManager   *StepManager
	streamManager *StreamManager
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
	}
	return m
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

	steps := []*Step{}
	for i, tS := range t.Steps {
		step := m.stepManager.create(b, i, &tS)

		for _, streamName := range tS.GetOutputStreams() {
			stream := m.streamManager.create(step, streamName)
			step.Streams = append(step.Streams, stream)
		}
		steps = append(steps, step)
	}
	b.Steps = steps

	m.db.save(b)

	return b, nil
}

func (m *BuildManager) Update(b *Build) error {
	return m.db.save(b)
}

func (m *BuildManager) GetBuildByID(id string) (*Build, error) {
	return GetBuildByID(m.db.DB, id)
}

func (m *BuildManager) GetAllForProject(p *project.Project, q *domain.PagingQuery) ([]*Build, int) {
	return m.db.getAllForProject(p, q)
}

func (m *BuildManager) GetAllForCommit(c *githistory.Commit, q *domain.PagingQuery) ([]*Build, int) {
	return m.db.getAllForCommit(c, q)
}

// func (m *BuildManager) GetAllForBranch(b *githistory.Branch, q *domain.PagingQuery) ([]*Build, int) {
// 	return m.db.getAllForBranch(b, q)
// }

func (m *BuildManager) GetAllForTask(t *task.Task, q *domain.PagingQuery) ([]*Build, int) {
	return m.db.getAllForTask(t, q)
}

func (m *BuildManager) GetRunningBuilds() ([]*Build, int) {
	return m.db.getRunningBuilds()
}

func (m *BuildManager) GetWaitingBuilds() ([]*Build, int) {
	return m.db.getWaitingBuilds()
}
