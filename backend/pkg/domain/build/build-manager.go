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
	StreamManager *StreamManager
}

func NewBuildManager(
	db *storm.DB,
	stepManager *StepManager,
	streamManager *StreamManager,
) *BuildManager {
	m := &BuildManager{
		db:            newBuildStormDB(db),
		stepManager:   stepManager,
		StreamManager: streamManager,
	}
	return m
}

func (m *BuildManager) New(
	t *task.Task,
	params map[string]string,
) *Build {
	timestamp := time.Now().UTC()
	b := &Build{
		UUID:       uuid.NewV3(uuid.NewV1(), t.UUID).String(),
		Task:       t,
		Parameters: params,
		CreatedAt:  timestamp,
		UpdatedAt:  timestamp,
		Status:     velocity.StateWaiting,
	}

	steps := []*Step{}
	for i, tS := range t.Steps {
		step := m.stepManager.new(b, i, &tS)
		m.stepManager.Save(step)

		for _, streamName := range tS.GetOutputStreams() {
			stream := m.StreamManager.new(step, streamName)
			step.Streams = append(step.Streams, stream)
			m.StreamManager.save(stream)
		}
		steps = append(steps, step)
	}
	b.Steps = steps

	return b
}

func (m *BuildManager) Save(b *Build) error {
	return m.db.save(b)
}

func (m *BuildManager) GetBuildByUUID(uuid string) (*Build, error) {
	return GetBuildByUUID(m.db.DB, uuid)
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
