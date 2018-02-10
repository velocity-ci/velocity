package task

import (
	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/docker/go/canonical/json"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type StormTask struct {
	ID       string `storm:"id"`
	CommitID string `storm:"index"`
	Slug     string `storm:"index"`
	Name     string
	VTask    []byte
}

func (g *StormTask) ToTask(db *storm.DB) *Task {
	vTask := velocity.Task{}
	if err := json.Unmarshal(g.VTask, &vTask); err != nil {
		logrus.Error(err)
	}
	c, err := githistory.GetCommitByID(db, g.CommitID)
	if err != nil {
		logrus.Error(err)
	}
	return &Task{
		ID:     g.ID,
		Slug:   g.Slug,
		VTask:  &vTask,
		Commit: c,
	}
}

func (t *Task) ToStormTask() *StormTask {
	jsonTask, err := json.Marshal(t.VTask)
	if err != nil {
		logrus.Error(err)
	}

	return &StormTask{
		ID:       t.ID,
		Slug:     t.Slug,
		CommitID: t.Commit.ID,
		VTask:    jsonTask,
	}
}

type stormDB struct {
	*storm.DB
}

func newStormDB(db *storm.DB) *stormDB {
	db.Init(&StormTask{})
	return &stormDB{db}
}

func (db *stormDB) save(t *Task) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	if err := tx.Save(t.ToStormTask()); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (db *stormDB) getByCommitAndSlug(commit *githistory.Commit, name string) (*Task, error) {
	query := db.Select(q.And(q.Eq("CommitID", commit.ID), q.Eq("Slug", name)))
	var t StormTask
	if err := query.First(&t); err != nil {
		return nil, err
	}

	return t.ToTask(db.DB), nil
}

func (db *stormDB) getAllForCommit(commit *githistory.Commit, pQ *domain.PagingQuery) (r []*Task, t int) {
	t = 0
	query := db.Select(q.Eq("CommitID", commit.ID))
	t, err := query.Count(&StormTask{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}
	var stormTasks []StormTask
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&stormTasks)

	for _, st := range stormTasks {
		r = append(r, st.ToTask(db.DB))
	}

	return r, t
}

func GetByID(db *storm.DB, id string) (*Task, error) {
	var sT StormTask
	if err := db.One("ID", id, &sT); err != nil {
		return nil, err
	}
	return sT.ToTask(db), nil
}
