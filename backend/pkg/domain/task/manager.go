package task

import (
	"fmt"

	"github.com/asdine/storm"
	"github.com/gosimple/slug"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type Manager struct {
	db             *stormDB
	projectManager *project.Manager
	branchManager  *githistory.BranchManager
	commitManager  *githistory.CommitManager
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
	}
	return m
}

func (m *Manager) New(
	c *githistory.Commit,
	vTask *velocity.Task,
	setupStep velocity.Step,
) *Task {
	vTask.Steps = append([]velocity.Step{setupStep}, vTask.Steps...)
	return &Task{
		UUID:   uuid.NewV3(uuid.NewV1(), c.UUID).String(),
		Commit: c,
		Task:   vTask,
		Slug:   slug.Make(vTask.Name),
	}
}

func (m *Manager) Save(t *Task) error {
	return m.db.save(t)
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
	if err := m.projectManager.Save(p); err != nil {
		return nil, err
	}

	go sync(p, m)

	return p, nil
}
