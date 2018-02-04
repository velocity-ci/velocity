package build

import (
	"encoding/json"
	"time"

	"github.com/velocity-ci/velocity/backend/velocity"

	"github.com/asdine/storm/q"

	"github.com/velocity-ci/velocity/backend/pkg/domain"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
)

type stormBuild struct {
	ID          string `storm:"id"`
	TaskID      string `storm:"index"`
	CommitID    string `storm:"index"`
	ProjectID   string `storm:"index"`
	Parameters  []byte
	Status      string
	Steps       []stormStep
	CreatedAt   time.Time
	UpdatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
}

func (s *stormBuild) toBuild(db *storm.DB) *Build {
	params := map[string]string{}
	err := json.Unmarshal(s.Parameters, &params)
	if err != nil {
		logrus.Error(err)
	}
	t, err := task.GetByUUID(db, s.TaskID)
	if err != nil {
		logrus.Error(err)
	}

	steps := []*Step{}
	for _, s := range s.Steps {
		steps = append(steps, s.toStep())
	}
	return &Build{
		UUID:        s.ID,
		Task:        t,
		Parameters:  params,
		Status:      s.Status,
		Steps:       steps,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
		StartedAt:   s.StartedAt,
		CompletedAt: s.CompletedAt,
	}
}

func (b *Build) toStormBuild() *stormBuild {
	paramsJson, err := json.Marshal(b.Parameters)
	if err != nil {
		logrus.Error(err)
	}
	steps := []stormStep{}
	for _, s := range b.Steps {
		steps = append(steps, s.toStormStep())
	}
	return &stormBuild{
		ID:          b.UUID,
		TaskID:      b.Task.UUID,
		CommitID:    b.Task.Commit.UUID,
		ProjectID:   b.Task.Commit.Project.UUID,
		Parameters:  paramsJson,
		Steps:       steps,
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
	db.Init(&stormBuild{})
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
	query := db.Select(q.Eq("ProjectID", p.UUID)).OrderBy("CreatedAt").Reverse()
	t, err := query.Count(&stormBuild{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}

	var stormBuilds []stormBuild
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&stormBuilds)

	for _, sB := range stormBuilds {
		r = append(r, sB.toBuild(db.DB))
	}

	return r, t
}

func (db *buildStormDB) getAllForCommit(c *githistory.Commit, pQ *domain.PagingQuery) (r []*Build, t int) {
	t = 0
	query := db.Select(q.Eq("CommitID", c.UUID)).OrderBy("CreatedAt").Reverse()
	t, err := query.Count(&stormBuild{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}

	var stormBuilds []stormBuild
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&stormBuilds)

	for _, sB := range stormBuilds {
		r = append(r, sB.toBuild(db.DB))
	}

	return r, t
}

func (db *buildStormDB) getAllForTask(tsk *task.Task, pQ *domain.PagingQuery) (r []*Build, t int) {
	t = 0
	query := db.Select(q.Eq("TaskID", tsk.UUID)).OrderBy("CreatedAt").Reverse()
	t, err := query.Count(&stormBuild{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}

	var stormBuilds []stormBuild
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&stormBuilds)

	for _, sB := range stormBuilds {
		r = append(r, sB.toBuild(db.DB))
	}

	return r, t
}

func (db *buildStormDB) getRunningBuilds() (r []*Build, t int) {
	t = 0
	query := db.Select(q.Eq("Status", velocity.StateRunning)).OrderBy("CreatedAt").Reverse()
	t, err := query.Count(&stormBuild{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}

	var stormBuilds []stormBuild
	query.Find(&stormBuilds)

	for _, sB := range stormBuilds {
		r = append(r, sB.toBuild(db.DB))
	}

	return r, t
}

func (db *buildStormDB) getWaitingBuilds() (r []*Build, t int) {
	t = 0
	query := db.Select(q.Eq("Status", velocity.StateWaiting)).OrderBy("CreatedAt").Reverse()
	t, err := query.Count(&stormBuild{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}

	var stormBuilds []stormBuild
	query.Find(&stormBuilds)

	for _, sB := range stormBuilds {
		r = append(r, sB.toBuild(db.DB))
	}

	return r, t
}

func GetBuildByUUID(db *storm.DB, uuid string) (*Build, error) {
	var sB stormBuild
	if err := db.One("ID", uuid, &sB); err != nil {
		logrus.Error(err)
		return nil, err
	}
	return sB.toBuild(db), nil
}
