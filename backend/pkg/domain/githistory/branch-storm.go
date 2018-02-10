package githistory

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type StormBranch struct {
	ID          string `storm:"id"`
	ProjectID   string `storm:"index"`
	Name        string
	LastUpdated time.Time
	Active      bool
}

func (s *StormBranch) ToBranch(db *storm.DB) *Branch {
	p, err := project.GetByID(db, s.ProjectID)
	if err != nil {
		logrus.Error(err)
	}
	return &Branch{
		ID:          s.ID,
		Project:     p,
		Name:        s.Name,
		LastUpdated: s.LastUpdated,
		Active:      s.Active,
	}
}

func (b *Branch) ToStormBranch() *StormBranch {
	return &StormBranch{
		ID:          b.ID,
		ProjectID:   b.Project.ID,
		Name:        b.Name,
		LastUpdated: b.LastUpdated,
		Active:      b.Active,
	}
}

type branchCommitStorm struct {
	ID       string `storm:"id"`
	BranchID string `storm:"index"`
	CommitID string `storm:"index"`
}

func newBranchCommitStorm(b *Branch, c *Commit) *branchCommitStorm {
	return &branchCommitStorm{
		ID:       fmt.Sprintf("%s:%s", b.Name, c.Hash),
		BranchID: b.ID,
		CommitID: c.ID,
	}
}

type branchStormDB struct {
	*storm.DB
}

func newBranchStormDB(db *storm.DB) *branchStormDB {
	db.Init(&Branch{})
	db.Init(&Commit{})
	return &branchStormDB{db}
}

func (db *branchStormDB) save(b *Branch) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	if err := tx.Save(b.ToStormBranch()); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (db *branchStormDB) getAllForProject(p *project.Project, pQ *domain.PagingQuery) (r []*Branch, t int) {
	t = 0
	query := db.Select(q.Eq("ProjectID", p.ID))
	t, err := query.Count(&StormBranch{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	var StormBranches []*StormBranch
	query.Find(&StormBranches)
	for _, b := range StormBranches {
		r = append(r, b.ToBranch(db.DB))
	}

	return r, t
}

func (db *branchStormDB) getAllForCommit(c *Commit, pQ *domain.PagingQuery) (r []*Branch, t int) {
	t = 0
	query := db.Select(q.Eq("CommitID", c.ID))
	t, err := query.Count(&branchCommitStorm{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}
	branchCommits := []branchCommitStorm{}
	query.Find(&branchCommits)
	branchIDs := []string{}
	for _, bC := range branchCommits {
		branchIDs = append(branchIDs, bC.BranchID)
	}

	query = db.Select(q.In("ID", branchIDs))
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	var StormBranches []*StormBranch
	query.Find(&StormBranches)
	for _, b := range StormBranches {
		r = append(r, b.ToBranch(db.DB))
	}

	return r, t
}

func (db *branchStormDB) hasCommit(b *Branch, c *Commit) bool {
	query := db.Select(q.And(q.Eq("CommitID", c.ID), q.Eq("BranchID", b.ID)))
	if err := query.First(&branchCommitStorm{}); err != nil {
		return false
	}

	return true
}

func GetBranchByID(db *storm.DB, id string) (*Branch, error) {
	var b StormBranch
	if err := db.One("ID", id, &b); err != nil {
		return nil, err
	}
	return b.ToBranch(db), nil
}

func (db *branchStormDB) getByProjectAndName(p *project.Project, name string) (*Branch, error) {
	query := db.Select(q.And(q.Eq("ProjectID", p.ID), q.Eq("Name", name)))
	var b StormBranch
	if err := query.First(&b); err != nil {
		return nil, err
	}

	return b.ToBranch(db.DB), nil
}
