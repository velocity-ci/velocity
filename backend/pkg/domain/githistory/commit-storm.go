package githistory

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type stormCommit struct {
	ID        string `storm:"id"`
	ProjectID string `storm:"index"`
	Hash      string `storm:"index"`
	Author    string
	CreatedAt time.Time
	Message   string
}

func (s *stormCommit) ToCommit(db *storm.DB) *Commit {
	p, err := project.GetByUUID(db, s.ProjectID)
	if err != nil {
		logrus.Error(err)
	}
	return &Commit{
		UUID:      s.ID,
		Project:   p,
		Hash:      s.Hash,
		Author:    s.Author,
		CreatedAt: s.CreatedAt,
		Message:   s.Message,
	}
}

func (c *Commit) ToStormCommit() *stormCommit {
	return &stormCommit{
		ID:        c.UUID,
		ProjectID: c.Project.UUID,
		Hash:      c.Hash,
		Author:    c.Author,
		CreatedAt: c.CreatedAt,
		Message:   c.Message,
	}
}

type commitStormDB struct {
	*storm.DB
}

func newCommitStormDB(db *storm.DB) *commitStormDB {
	db.Init(&Branch{})
	db.Init(&Commit{})
	return &commitStormDB{db}
}

func (db *commitStormDB) getByProjectAndHash(p *project.Project, hash string) (*Commit, error) {
	query := db.Select(q.And(q.Eq("ProjectID", p.UUID), q.Eq("Hash", hash)))
	var c stormCommit
	if err := query.First(&c); err != nil {
		return nil, err
	}

	return c.ToCommit(db.DB), nil
}

func (db *commitStormDB) getAllForProject(p *project.Project, pQ *domain.PagingQuery) (r []*Commit, t int) {
	t = 0
	query := db.Select(q.Eq("ProjectID", p.UUID))
	t, err := query.Count(&stormCommit{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	var stormCommits []*stormCommit
	query.Find(&stormCommits)
	for _, c := range stormCommits {
		r = append(r, c.ToCommit(db.DB))
	}

	return r, t
}

func (db *commitStormDB) getAllForBranch(b *Branch, pQ *domain.PagingQuery) (r []*Commit, t int) {
	t = 0

	query := db.Select(q.Eq("BranchID", b.UUID))
	t, err := query.Count(&branchCommitStorm{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}
	branchCommits := []branchCommitStorm{}
	query.Find(&branchCommits)
	commitIDs := []string{}
	for _, bC := range branchCommits {
		commitIDs = append(commitIDs, bC.CommitID)
	}

	query = db.Select(q.In("ID", commitIDs))
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	var stormCommits []*stormCommit
	query.Find(&stormCommits)
	for _, c := range stormCommits {
		r = append(r, c.ToCommit(db.DB))
	}

	return r, t
}

func GetCommitByUUID(db *storm.DB, uuid string) (*Commit, error) {
	var c stormCommit
	if err := db.One("ID", uuid, &c); err != nil {
		return nil, err
	}
	return c.ToCommit(db), nil
}
