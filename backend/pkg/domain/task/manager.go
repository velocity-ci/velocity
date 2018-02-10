package task

import (
	"fmt"

	"github.com/asdine/storm"
	"github.com/gosimple/slug"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

// Event constants
const (
	EventCreate = "task:new"
)

type Manager struct {
	db             *stormDB
	projectManager *project.Manager
	branchManager  *githistory.BranchManager
	commitManager  *githistory.CommitManager
	brokers        []domain.Broker
}

func NewManager(
	db *storm.DB,
	projectManager *project.Manager,
	branchManager *githistory.BranchManager,
	commitManager *githistory.CommitManager,
) *Manager {
	m := &Manager{
		db:             newStormDB(db),
		projectManager: projectManager,
		branchManager:  branchManager,
		commitManager:  commitManager,
		brokers:        []domain.Broker{},
	}
	return m
}

func (m *Manager) AddBroker(b domain.Broker) {
	m.brokers = append(m.brokers, b)
}

func (m *Manager) Create(
	c *githistory.Commit,
	vTask *velocity.Task,
	setupStep velocity.Step,
) *Task {
	vTask.Steps = append([]velocity.Step{setupStep}, vTask.Steps...)
	t := &Task{
		ID:     uuid.NewV3(uuid.NewV1(), c.ID).String(),
		Commit: c,
		VTask:  vTask,
		Slug:   slug.Make(vTask.Name),
	}

	m.db.save(t)

	// for _, b := range m.brokers {
	// 	b.EmitAll(&domain.Emit{
	// 		Event:   EventCreate,
	// 		Payload: t,
	// 	})

	// }

	return t
}

func (m *Manager) GetByCommitAndSlug(c *githistory.Commit, slug string) (*Task, error) {
	return m.db.getByCommitAndSlug(c, slug)
}

func (m *Manager) GetAllForCommit(c *githistory.Commit, q *domain.PagingQuery) ([]*Task, int) {
	return m.db.getAllForCommit(c, q)
}

func (m *Manager) Sync(p *project.Project) (*project.Project, error) {
	if p.Synchronising {
		return nil, fmt.Errorf("already synchronising")
	}

	p.Synchronising = true
	if err := m.projectManager.Update(p); err != nil {
		return nil, err
	}

	go sync(p, m)

	return p, nil
}
