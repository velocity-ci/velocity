package githistory

import (
	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type commitStormDB struct {
	*storm.DB
}

func newCommitStormDB(db *storm.DB) *commitStormDB {
	db.Init(&Branch{})
	db.Init(&Commit{})
	return &commitStormDB{db}
}

func (db *commitStormDB) getByProjectAndHash(p *project.Project, hash string) (*Commit, error) {
	query := db.Select(q.And(q.Eq("Project", p), q.Eq("Hash", hash)))
	var c Commit
	if err := query.First(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (db *commitStormDB) getAllForProject(p *project.Project, pQ *domain.PagingQuery) (r []*Commit, t int) {
	t = 0
	query := db.Select(q.Eq("Project", p))
	t, err := query.Count(&Commit{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&r)

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

	query = db.Select(q.In("UUID", commitIDs))
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&r)

	return r, t
}

func GetCommitByUUID(db *storm.DB, uuid string) (*Commit, error) {
	var c Commit
	if err := db.One("UUID", uuid, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
