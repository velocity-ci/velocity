package githistory

import (
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type CommitManager struct {
	db *commitDB
}

func NewCommitManager(
	db *gorm.DB,
) *CommitManager {
	db.AutoMigrate(&GormCommit{}, &GormBranch{})
	m := &CommitManager{
		db: newCommitDB(db),
	}
	return m
}

func (m *CommitManager) New(
	p *project.Project,
	hash string,
	message string,
	author string,
	date time.Time,
	branches []*Branch,
) *Commit {
	return &Commit{
		UUID:      uuid.NewV3(uuid.NewV1(), p.UUID).String(),
		Project:   p,
		Hash:      hash,
		Message:   message,
		Author:    author,
		CreatedAt: date.UTC(),
		Branches:  branches,
	}
}

// func (m *CommitManager) Exists(p *project.Project, hash string) bool {
// 	if _, err := m.GetByHash(hash); err != nil {
// 		return false
// 	}
// 	return true
// }

func (m *CommitManager) Save(c *Commit) error {
	return m.db.save(c)
}

func (m *CommitManager) Delete(c *Commit) error {
	return m.db.delete(c)
}

// func (m *CommitManager) GetAllForProject(p *project.Project, q *domain.PagingQuery) ([]*Commit, int) {
// 	return m.db.getAll(q)
// }

func (m *CommitManager) GetByProjectAndHash(p *project.Project, hash string) (*Commit, error) {
	return m.db.getByProjectAndHash(p, hash)
}
