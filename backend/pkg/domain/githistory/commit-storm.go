package githistory

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type StormCommit struct {
	ID        string `storm:"id"`
	ProjectID string `storm:"index"`
	Hash      string `storm:"index"`
	Author    string
	CreatedAt time.Time
	Message   string
}

func (s *StormCommit) ToCommit(db *storm.DB) *Commit {
	p, err := project.GetByID(db, s.ProjectID)
	if err != nil {
		logrus.Error(err)
	}
	return &Commit{
		ID:        s.ID,
		Project:   p,
		Hash:      s.Hash,
		Author:    s.Author,
		CreatedAt: s.CreatedAt,
		Message:   s.Message,
	}
}

func (c *Commit) ToStormCommit() *StormCommit {
	return &StormCommit{
		ID:        c.ID,
		ProjectID: c.Project.ID,
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

func (db *commitStormDB) saveCommitToBranch(c *Commit, b *Branch) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	if err := tx.Save(c.ToStormCommit()); err != nil {
		tx.Rollback()
		return err
	}

	bC := newBranchCommitStorm(b, c)
	if err := tx.Save(bC); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (db *commitStormDB) getByProjectAndHash(p *project.Project, hash string) (*Commit, error) {
	query := db.Select(q.And(q.Eq("ProjectID", p.ID), q.Eq("Hash", hash)))
	var c StormCommit
	if err := query.First(&c); err != nil {
		return nil, err
	}

	return c.ToCommit(db.DB), nil
}

func (db *commitStormDB) getAllForProject(p *project.Project, pQ *CommitQuery) (r []*Commit, t int) {
	if len(pQ.Branches) > 0 {
		return db.getAllForProjectBranchFilter(p, pQ)
	}
	t = 0
	query := db.Select(q.Eq("ProjectID", p.ID))
	t, err := query.Count(&StormCommit{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	var stormCommits []*StormCommit
	query.Find(&stormCommits)
	for _, dC := range stormCommits {
		r = append(r, dC.ToCommit(db.DB))
	}

	return r, t
}

func (db *commitStormDB) getAllForProjectBranchFilter(p *project.Project, pQ *CommitQuery) (r []*Commit, t int) {
	t = 0
	skipCounter := 0
	query := db.Select(q.Eq("ProjectID", p.ID))
	var stormCommits []*StormCommit
	query.Find(&stormCommits)
	for _, dC := range stormCommits {
		// if unison of branches add
		query := db.Select(q.Eq("CommitID", dC.ID))
		branchCommits := []branchCommitStorm{}
		query.Find(&branchCommits)
		branchIDs := []string{}
		for _, bC := range branchCommits {
			branchIDs = append(branchIDs, bC.BranchID)
		}
		query = db.Select(q.In("ID", branchIDs))
		var stormBranches []*StormBranch
		query.Find(&stormBranches)
		if isStormBranchInBranchNames(stormBranches, pQ.Branches) {
			t++
			if len(r) >= pQ.Limit {
				break
			} else if skipCounter < (pQ.Page-1)*pQ.Limit {
				skipCounter++
			} else {
				r = append(r, dC.ToCommit(db.DB))
			}
		}
	}

	return r, t
}

func isStormBranchInBranchNames(sBs []*StormBranch, branches []string) bool {
	for _, sB := range sBs {
		for _, branch := range branches {
			if sB.Name == branch {
				return true
			}
		}
	}

	return false
}

func (db *commitStormDB) getAllForBranch(b *Branch, pQ *domain.PagingQuery) (r []*Commit, t int) {
	t = 0

	query := db.Select(q.Eq("BranchID", b.ID))
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
	var stormCommits []*StormCommit
	query.Find(&stormCommits)
	for _, c := range stormCommits {
		r = append(r, c.ToCommit(db.DB))
	}

	return r, t
}

func GetCommitByID(db *storm.DB, id string) (*Commit, error) {
	var c StormCommit
	if err := db.One("ID", id, &c); err != nil {
		return nil, err
	}
	return c.ToCommit(db), nil
}
