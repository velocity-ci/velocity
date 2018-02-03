package task

import (
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type Manager struct {
	db *db
}

func NewManager(
	db *gorm.DB,
) *Manager {
	db.AutoMigrate(&GormTask{})
	m := &Manager{
		db: newDB(db),
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
	}
}

func (m *Manager) Save(t *Task) error {
	return m.db.save(t)
}

func (m *Manager) GetByCommitAndName(c *githistory.Commit, name string) (*Task, error) {
	return m.db.getByCommitAndName(c, name)
}

func (m *Manager) GetAllForCommit(c *githistory.Commit, q *domain.PagingQuery) ([]*Task, int) {
	return m.db.getAllForCommit(c, q)
}
