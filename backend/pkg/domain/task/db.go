package task

// import (
// 	"fmt"

// 	"github.com/Sirupsen/logrus"
// 	"github.com/docker/go/canonical/json"
// 	"github.com/jinzhu/gorm"
// 	"github.com/velocity-ci/velocity/backend/pkg/domain"
// 	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
// 	"github.com/velocity-ci/velocity/backend/velocity"
// )

// type GormTask struct {
// 	UUID string `gorm:"primary_key"`
// 	// Commit   *githistory.GormCommit `gorm:"ForeignKey:CommitID"`
// 	CommitID string
// 	Name     string `gorm:"not null"`
// 	VTask    []byte
// }

// func (GormTask) TableName() string {
// 	return "tasks"
// }

// func (g *GormTask) ToTask() *Task {
// 	vTask := velocity.Task{}
// 	err := json.Unmarshal(g.VTask, &vTask)
// 	if err != nil {
// 		logrus.Error(err)
// 	}
// 	return &Task{
// 		UUID: g.UUID,
// 		// Commit: g.Commit.ToCommit(),
// 		Task: &vTask,
// 	}
// }

// func (t *Task) ToGormTask() *GormTask {
// 	jsonTask, err := json.Marshal(t.Task)
// 	if err != nil {
// 		logrus.Error(err)
// 	}

// 	return &GormTask{
// 		UUID: t.UUID,
// 		// Commit: t.Commit.ToGormCommit(),
// 		Name:  t.Name,
// 		VTask: jsonTask,
// 	}
// }

// type db struct {
// 	db *gorm.DB
// }

// func newDB(gorm *gorm.DB) *db {
// 	gorm.AutoMigrate(&GormTask{})

// 	return &db{
// 		db: gorm,
// 	}
// }

// func (db *db) save(t *Task) error {
// 	tx := db.db.Begin()

// 	g := t.ToGormTask()

// 	tx.
// 		Where(GormTask{UUID: t.UUID}).
// 		Assign(&g).
// 		FirstOrCreate(&g)

// 	return tx.Commit().Error
// }

// func (db *db) getByCommitAndName(commit *githistory.Commit, name string) (*Task, error) {
// 	g := GormTask{}
// 	if db.db.
// 		Preload("Commit").
// 		Preload("Commit.Project").
// 		Where("commit_id = ? AND name = ?", commit.UUID, name).
// 		First(&g).RecordNotFound() {
// 		return nil, fmt.Errorf("could not find commit:task %s:%s", commit.UUID, name)
// 	}
// 	return g.ToTask(), nil
// }

// func (db *db) getAllForCommit(commit *githistory.Commit, q *domain.PagingQuery) (r []*Task, t int) {
// 	t = 0
// 	gS := []GormTask{}
// 	d := db.db

// 	d = d.
// 		Preload("Commit").
// 		Preload("Commit.Project").
// 		Where("commit_id = ?", commit.UUID).
// 		Find(&gS).
// 		Count(&t)

// 	d.
// 		Limit(q.Limit).
// 		Offset((q.Page - 1) * q.Limit).
// 		Find(&gS)

// 	for _, g := range gS {
// 		r = append(r, g.ToTask())
// 	}

// 	return r, t
// }
