package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/velocity-ci/velocity/backend/api/commit"
	"github.com/velocity-ci/velocity/backend/api/project"
	"github.com/velocity-ci/velocity/backend/task"
)

type Manager struct {
	logger *log.Logger
	bolt   *bolt.DB
}

func NewManager(
	bolt *bolt.DB,
) *Manager {
	return &Manager{
		logger: log.New(os.Stdout, "[bolt-task]", log.Lshortfile),
		bolt:   bolt,
	}
}

func (m *Manager) SaveTaskForCommitInProject(t *task.Task, c *commit.Commit, p *project.Project) error {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	if projectsBucket == nil {
		return errors.New("Projects not found:")
	}
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
	if projectBucket == nil {
		return fmt.Errorf("Project not found: %s", p.ID)
	}

	commitsBucket := projectBucket.Bucket([]byte("commits"))
	if commitsBucket == nil {
		return fmt.Errorf("Could not find any commits for project: %s", p.ID)
	}

	commitBucket := commitsBucket.Bucket([]byte(c.Hash))
	if commitBucket == nil {
		return fmt.Errorf("Could not find project: %s, commit: %s", p.ID, c.Hash)
	}

	tasksBucket, err := commitBucket.CreateBucketIfNotExists([]byte("tasks"))
	if err != nil {
		return err
	}
	if tasksBucket == nil {
		tasksBucket = commitBucket.Bucket([]byte("tasks"))
	}

	taskJSON, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
	}
	tasksBucket.Put([]byte(t.Name), taskJSON)

	if err := tx.Commit(); err != nil {
		return err
	}

	m.logger.Printf("Saved task %s for %s in %s", t.Name, c.Hash, p.ID)

	return nil
}

func (m *Manager) GetTasksForCommitInProject(c *commit.Commit, p *project.Project) []task.Task {
	tasks := []task.Task{}

	tx, err := m.bolt.Begin(false)
	if err != nil {
		return tasks
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	if projectsBucket == nil {
		return tasks
	}
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
	if projectBucket == nil {
		return tasks
	}

	commitsBucket := projectBucket.Bucket([]byte("commits"))
	if commitsBucket == nil {
		return tasks
	}

	commitBucket := commitsBucket.Bucket([]byte(c.Hash))
	if commitBucket == nil {
		return tasks
	}

	tasksBucket := commitBucket.Bucket([]byte("tasks"))
	if tasksBucket == nil {
		return tasks
	}

	cursor := tasksBucket.Cursor()
	for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
		task := task.NewTask()
		err := json.Unmarshal(v, &task)
		if err == nil {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

func (m *Manager) GetTaskForCommitInProject(c *commit.Commit, p *project.Project, name string) (*task.Task, error) {

	tx, err := m.bolt.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	if projectsBucket == nil {
		return nil, fmt.Errorf("Could not find commit for project: %s", p.ID)
	}

	projectBucket := projectsBucket.Bucket([]byte(p.ID))
	if projectsBucket == nil {
		return nil, fmt.Errorf("Could not find commit for project: %s", p.ID)
	}

	commitsBucket := projectBucket.Bucket([]byte("commits"))
	if commitsBucket == nil {
		return nil, fmt.Errorf("Could not find commit for project: %s", p.ID)
	}

	commitBucket := commitsBucket.Bucket([]byte(c.Hash))
	if commitBucket == nil {
		return nil, fmt.Errorf("Could not find commit for project: %s", p.ID)
	}

	tasksBucket := commitBucket.Bucket([]byte("tasks"))
	if tasksBucket == nil {
		return nil, fmt.Errorf("Could not find commit for project: %s", p.ID)
	}

	t := tasksBucket.Get([]byte(name))

	task := task.NewTask()
	err = json.Unmarshal(t, &task)

	if err != nil {
		return nil, fmt.Errorf("Could not find commit for project: %s", p.ID)
	}

	return &task, nil

}
