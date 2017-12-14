package task

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type gormTask struct {
	ID       string `gorm:"primary_key"`
	CommitID string
	Name     string
	VTask    []byte // JSON of task for storage
}

func (gormTask) TableName() string {
	return "tasks"
}

func gormTaskFromTask(t Task) gormTask {
	jsonTask, err := json.Marshal(t.Task)
	if err != nil {
		log.Println("Could not marshal task")
		log.Fatal(err)
	}

	return gormTask{
		ID:       t.ID,
		CommitID: t.CommitID,
		Name:     t.Name,
		VTask:    jsonTask,
	}
}

func taskFromGormTask(g gormTask) Task {
	var vTask velocity.Task
	err := json.Unmarshal(g.VTask, &vTask)
	if err != nil {
		log.Println("Could not unmarshal task from %v", g.VTask)
		log.Fatal(err)
	}

	return Task{
		ID:       g.ID,
		CommitID: g.CommitID,
		Task:     vTask,
	}
}

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	logger *log.Logger
	gorm   *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(gormTask{})
	return &gormRepository{
		logger: log.New(os.Stdout, "[gorm:task]", log.Lshortfile),
		gorm:   db,
	}
}

func (r *gormRepository) Save(t Task) Task {
	tx := r.gorm.Begin()

	gT := gormTaskFromTask(t)

	err := tx.Where(&gormTask{
		ID: t.ID,
	}).First(&gormTask{}).Error
	if err != nil {
		err = tx.Create(&gT).Error
	} else {
		tx.Save(&gT)
	}

	tx.Commit()
	r.logger.Printf("saved task %s", t.ID)
	return taskFromGormTask(gT)
}

func (r *gormRepository) Delete(t Task) {
	tx := r.gorm.Begin()

	gT := gormTaskFromTask(t)

	if err := tx.Delete(gT).Error; err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	tx.Commit()
}

func (r *gormRepository) GetByTaskID(taskID string) (Task, error) {
	gT := gormTask{}

	if r.gorm.
		Where(&gormTask{
			ID: taskID,
		}).
		First(&gT).RecordNotFound() {
		log.Printf("could not find task %s", taskID)
		return Task{}, fmt.Errorf("could not find task %s", taskID)
	}

	return taskFromGormTask(gT), nil
}

func (r *gormRepository) GetByCommitIDAndTaskName(commitID string, name string) (Task, error) {
	gT := gormTask{}

	if r.gorm.
		Where(&gormTask{
			CommitID: commitID,
			Name:     name,
		}).
		First(&gT).RecordNotFound() {
		log.Printf("could not find commit:task %s:%s", commitID, name)
		return Task{}, fmt.Errorf("could not find commit:task %s:%s", commitID, name)
	}

	return taskFromGormTask(gT), nil
}

func (r *gormRepository) GetAllByCommitID(commitID string, q Query) ([]Task, uint64) {
	gTs := []gormTask{}
	var count uint64

	r.gorm.
		Where(&gormTask{
			CommitID: commitID,
		}).
		Find(&gTs).
		Count(&count)

	tasks := []Task{}
	for _, gT := range gTs {
		tasks = append(tasks, taskFromGormTask(gT))
	}

	return tasks, count
}
