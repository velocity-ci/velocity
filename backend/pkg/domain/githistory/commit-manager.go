package githistory

import (
	"time"

	"github.com/asdine/storm"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type CommitManager struct {
	db *commitStormDB
}

func NewCommitManager(
	db *storm.DB,
) *CommitManager {
	m := &CommitManager{
		db: newCommitStormDB(db),
	}
	return m
}

func (m *CommitManager) Create(
	b *Branch,
	p *project.Project,
	hash string,
	message string,
	author string,
	date time.Time,
) *Commit {
	c := &Commit{
		ID:        uuid.NewV3(uuid.NewV1(), p.ID).String(),
		Project:   p,
		Hash:      hash,
		Message:   message,
		Author:    author,
		CreatedAt: date.UTC(),
	}

	m.db.saveCommitToBranch(c, b)
	return c
}

func (m *CommitManager) AddCommitToBranch(c *Commit, b *Branch) error {
	return m.db.saveCommitToBranch(c, b)
}

func (m *CommitManager) GetAllForProject(p *project.Project, q *domain.PagingQuery) ([]*Commit, int) {
	return m.db.getAllForProject(p, q)
}

func (m *CommitManager) GetAllForBranch(b *Branch, q *domain.PagingQuery) ([]*Commit, int) {
	return m.db.getAllForBranch(b, q)
}

func (m *CommitManager) GetByProjectAndHash(p *project.Project, hash string) (*Commit, error) {
	return m.db.getByProjectAndHash(p, hash)
}
