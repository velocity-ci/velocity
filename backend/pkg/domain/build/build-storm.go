package build

import (
	"encoding/json"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/velocity"

	"github.com/asdine/storm/q"

	"github.com/velocity-ci/velocity/backend/pkg/domain"

	"github.com/asdine/storm"
	"github.com/golang/glog"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
)

type StormBuild struct {
	ID          string `storm:"id"`
	TaskID      string `storm:"index"`
	CommitID    string `storm:"index"`
	ProjectID   string `storm:"index"`
	Parameters  []byte
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
}

func (s *StormBuild) toBuild(db *storm.DB) *Build {
	params := map[string]string{}
	err := json.Unmarshal(s.Parameters, &params)
	if err != nil {
		glog.Error(err)
	}
	t, err := task.GetByID(db, s.TaskID)
	if err != nil {
		glog.Error(err)
	}

	return &Build{
		ID:          s.ID,
		Task:        t,
		Parameters:  params,
		Status:      s.Status,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		StartedAt:   s.StartedAt,
		CompletedAt: s.CompletedAt,
	}
}

func (b *Build) toStormBuild() *StormBuild {
	paramsJson, err := json.Marshal(b.Parameters)
	if err != nil {
		glog.Error(err)
	}
	return &StormBuild{
		ID:          b.ID,
		TaskID:      b.Task.ID,
		CommitID:    b.Task.Commit.ID,
		ProjectID:   b.Task.Commit.Project.ID,
		Parameters:  paramsJson,
		Status:      b.Status,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
		StartedAt:   b.StartedAt,
		CompletedAt: b.CompletedAt,
	}
}

type buildStormDB struct {
	*storm.DB
}

func newBuildStormDB(db *storm.DB) *buildStormDB {
	db.Init(&StormBuild{})
	return &buildStormDB{db}
}

func (db *buildStormDB) save(b *Build) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	if err := tx.Save(b.toStormBuild()); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (db *buildStormDB) getAllForProject(p *project.Project, pQ *domain.PagingQuery) (r []*Build, t int) {
	t = 0
	query := db.Select(q.Eq("ProjectID", p.ID)).OrderBy("CreatedAt").Reverse()
	t, err := query.Count(&StormBuild{})
	if err != nil {
		glog.Error(err)
		return r, t
	}

	var StormBuilds []StormBuild
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&StormBuilds)

	for _, sB := range StormBuilds {
		r = append(r, sB.toBuild(db.DB))
	}

	return r, t
}

func (db *buildStormDB) getAllForCommit(c *githistory.Commit, pQ *domain.PagingQuery) (r []*Build, t int) {
	t = 0
	query := db.Select(q.Eq("CommitID", c.ID)).OrderBy("CreatedAt").Reverse()
	t, err := query.Count(&StormBuild{})
	if err != nil {
		glog.Error(err)
		return r, t
	}

	var StormBuilds []StormBuild
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&StormBuilds)

	for _, sB := range StormBuilds {
		r = append(r, sB.toBuild(db.DB))
	}

	return r, t
}

func (db *buildStormDB) getAllForTask(tsk *task.Task, pQ *domain.PagingQuery) (r []*Build, t int) {
	t = 0
	query := db.Select(q.Eq("TaskID", tsk.ID)).OrderBy("CreatedAt").Reverse()
	t, err := query.Count(&StormBuild{})
	if err != nil {
		glog.Error(err)
		return r, t
	}

	var StormBuilds []StormBuild
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&StormBuilds)

	for _, sB := range StormBuilds {
		r = append(r, sB.toBuild(db.DB))
	}

	return r, t
}

func (db *buildStormDB) getRunningBuilds() (r []*Build, t int) {
	t = 0
	query := db.Select(q.Eq("Status", velocity.StateRunning)).OrderBy("CreatedAt").Reverse()
	t, err := query.Count(&StormBuild{})
	if err != nil {
		glog.Error(err)
		return r, t
	}

	var StormBuilds []StormBuild
	query.Find(&StormBuilds)

	for _, sB := range StormBuilds {
		r = append(r, sB.toBuild(db.DB))
	}

	return r, t
}

func (db *buildStormDB) getWaitingBuilds() (r []*Build, t int) {
	t = 0
	query := db.Select(q.Eq("Status", velocity.StateWaiting)).OrderBy("CreatedAt").Reverse()
	t, err := query.Count(&StormBuild{})
	if err != nil {
		glog.Error(err)
		return r, t
	}

	var StormBuilds []StormBuild
	query.Find(&StormBuilds)

	for _, sB := range StormBuilds {
		r = append(r, sB.toBuild(db.DB))
	}

	return r, t
}

func GetBuildByID(db *storm.DB, id string) (*Build, error) {
	var sB StormBuild
	if err := db.One("ID", id, &sB); err != nil {
		glog.Error(err)
		return nil, err
	}
	return sB.toBuild(db), nil
}
