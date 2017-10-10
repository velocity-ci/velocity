package commit

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/domain/task"
)

type BoltManager struct {
	logger *log.Logger
	bolt   *bolt.DB
}

func NewBoltManager(
	bolt *bolt.DB,
) *BoltManager {
	return &BoltManager{
		logger: log.New(os.Stdout, "[bolt-commit]", log.Lshortfile),
		bolt:   bolt,
	}
}

func (m *BoltManager) GetCommitInProject(hash string, p *domain.Project) (*domain.Commit, error) {
	tx, err := m.bolt.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	if projectsBucket == nil {
		return nil, errors.New("Projects not found:")
	}
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
	if projectBucket == nil {
		return nil, fmt.Errorf("Project not found: %s", p.ID)
	}

	commitsBucket := projectBucket.Bucket([]byte("commits"))
	if commitsBucket == nil {
		return nil, fmt.Errorf("Could not find any commits for project: %s", p.ID)
	}

	cursor := commitsBucket.Cursor()
	for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {

		key := string(k)

		if key[len(key)-7:] == hash[:7] {
			commitBucket := commitsBucket.Bucket(k)
			v := commitBucket.Get([]byte("info"))

			c := domain.Commit{}
			err = json.Unmarshal(v, &c)
			if err != nil {
				return nil, err
			}

			return &c, nil
		}
	}

	return nil, fmt.Errorf("Could not find project: %s, commit: %s", p.ID, hash)
}

type CommitQueryOpts struct {
	Branch string
	Amount int
	Page   int
}

func (m *BoltManager) FindAllCommitsForProject(p *domain.Project, queryOpts *CommitQueryOpts) []domain.Commit {
	commits := []domain.Commit{}

	tx, err := m.bolt.Begin(false)
	if err != nil {
		return commits
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
	commitsBucket := projectBucket.Bucket([]byte("commits"))
	if commitsBucket == nil {
		return commits
	}

	c := commitsBucket.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		cB := commitsBucket.Bucket(k)
		v := cB.Get([]byte("info"))
		commit := domain.Commit{}
		err := json.Unmarshal(v, &commit)
		if err == nil {
			commits = append(commits, commit)
		}
	}

	return commits
}

func (m *BoltManager) SaveCommitForProject(p *domain.Project, c *domain.Commit) error {
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
		commitsBucket, err = projectBucket.CreateBucket([]byte("commits"))
		if err != nil {
			return err
		}
	}

	commitBucket, err := commitsBucket.CreateBucketIfNotExists([]byte(c.OrderedID()))
	if err != nil {
		return err
	}
	if commitBucket == nil {
		commitBucket = commitsBucket.Bucket([]byte(c.OrderedID()))
	}

	commitJSON, _ := json.Marshal(c)
	commitBucket.Put([]byte("info"), commitJSON)

	if err := tx.Commit(); err != nil {
		return err
	}

	m.logger.Printf("Saved commit %s for %s", c.Hash, p.ID)

	return nil
}

func (m *BoltManager) SaveTaskForCommitInProject(t *task.Task, c *domain.Commit, p *domain.Project) error {
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

func (m *BoltManager) GetTasksForCommitInProject(c *domain.Commit, p *domain.Project) []task.Task {
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

func (m *BoltManager) GetTaskForCommitInProject(c *domain.Commit, p *domain.Project, name string) (*task.Task, error) {

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
