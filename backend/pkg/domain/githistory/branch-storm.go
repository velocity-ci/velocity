package githistory

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type branchCommitStorm struct {
	ID       string `storm:"id"`
	BranchID string `storm:"index"`
	CommitID string `storm:"index"`
}

func newBranchCommitStorm(b *Branch, c *Commit) *branchCommitStorm {
	return &branchCommitStorm{
		ID:       fmt.Sprintf("%s:%s", b.Name, c.Hash),
		BranchID: b.UUID,
		CommitID: c.UUID,
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

	if err := tx.Save(b); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (db *branchStormDB) getAllForProject(p *project.Project, pQ *domain.PagingQuery) (r []*Branch, t int) {
	t = 0
	query := db.Select(q.Eq("Project", p))
	t, err := query.Count(&Branch{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&r)

	return r, t
}

func (db *branchStormDB) getAllForCommit(c *Commit, pQ *domain.PagingQuery) (r []*Branch, t int) {
	t = 0
	query := db.Select(q.Eq("CommitID", c.UUID))
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

	query = db.Select(q.In("UUID", branchIDs))
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&r)

	return r, t
}

func GetBranchByUUID(db *storm.DB, uuid string) (*Branch, error) {
	var b Branch
	if err := db.One("UUID", uuid, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func (db *branchStormDB) saveCommitToBranch(c *Commit, b *Branch) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	if err := tx.Save(c); err != nil {
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

func (db *branchStormDB) getByProjectAndName(p *project.Project, name string) (*Branch, error) {
	query := db.Select(q.And(q.Eq("Project", p), q.Eq("Name", name)))
	var b Branch
	if err := query.First(&b); err != nil {
		return nil, err
	}

	return &b, nil
}
