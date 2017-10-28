package commit

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
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
		logger: log.New(os.Stdout, "[bolt-commit]", log.Lshortfile),
		bolt:   bolt,
	}
}

func (m *Manager) GetCommitInProject(hash string, p *project.Project) (*Commit, error) {
	tx, err := m.bolt.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
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

			c := Commit{}
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

func (m *Manager) FindAllCommitsForProject(p *project.Project, queryOpts *CommitQueryOpts) []Commit {
	commits := []Commit{}

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

	skipCounter := 0
	c := commitsBucket.Cursor()
	for k, _ := c.Last(); k != nil; k, _ = c.Prev() {
		cB := commitsBucket.Bucket(k)
		v := cB.Get([]byte("info"))
		commit := Commit{}
		err := json.Unmarshal(v, &commit)
		if err == nil && (queryOpts.Branch == "" || commit.Branch == queryOpts.Branch) {
			if skipCounter < (queryOpts.Page-1)*queryOpts.Amount {
				skipCounter++
			} else {
				commits = append(commits, commit)
			}
		}
		if len(commits) >= queryOpts.Amount {
			break
		}
	}

	return commits
}

func (m *Manager) FindAllBranchesForProject(p *project.Project) []string {
	branches := []string{}

	tx, err := m.bolt.Begin(false)
	if err != nil {
		return branches
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
	branchesBucket := projectBucket.Bucket([]byte("branches"))
	if branchesBucket == nil {
		return branches
	}

	c := branchesBucket.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		branches = append(branches, string(k))
	}

	return branches
}

func (m *Manager) SaveBranchForProject(p *project.Project, branch string) error {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
	branchesBucket := projectBucket.Bucket([]byte("branches"))
	if branchesBucket == nil {
		branchesBucket, err = projectBucket.CreateBucket([]byte("branches"))
		if err != nil {
			return err
		}
	}

	branchesBucket.Put([]byte(branch), nil)

	return tx.Commit()
}

func (m *Manager) SaveCommitForProject(p *project.Project, c *Commit) error {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
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

	m.logger.Printf("Saving commit %s for %s", c.Hash, p.ID)

	return tx.Commit()
}

func (m *Manager) SaveTaskForCommitInProject(t *task.Task, c *Commit, p *project.Project) error {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
	commitsBucket := projectBucket.Bucket([]byte("commits"))
	commitBucket := commitsBucket.Bucket([]byte(c.OrderedID()))

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

func (m *Manager) GetTasksForCommitInProject(c *Commit, p *project.Project) []task.Task {
	tasks := []task.Task{}

	tx, err := m.bolt.Begin(false)
	if err != nil {
		return tasks
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
	commitsBucket := projectBucket.Bucket([]byte("commits"))
	commitBucket := commitsBucket.Bucket([]byte(c.OrderedID()))
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

func (m *Manager) GetTaskForCommitInProject(c *Commit, p *project.Project, name string) (*task.Task, error) {

	tx, err := m.bolt.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
	commitsBucket := projectBucket.Bucket([]byte("commits"))
	commitBucket := commitsBucket.Bucket([]byte(c.OrderedID()))

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

func (m *Manager) GetTotalCommitsForProject(p *project.Project) uint {
	var count uint
	count = 0

	tx, err := m.bolt.Begin(false)
	if err != nil {
		return count
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
	commitsBucket := projectBucket.Bucket([]byte("commits"))
	if commitsBucket == nil {
		return count
	}

	c := commitsBucket.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		count++
	}

	return count
}

func (m *Manager) SaveBuild(b *Build, p *project.Project, c *Commit) error {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
	commitsBucket := projectBucket.Bucket([]byte("commits"))
	commitBucket := commitsBucket.Bucket([]byte(c.OrderedID()))

	buildsBucket, err := commitBucket.CreateBucketIfNotExists([]byte("builds"))
	if err != nil {
		return err
	}
	if buildsBucket == nil {
		buildsBucket = commitBucket.Bucket([]byte("builds"))
	}

	buildJSON, err := json.Marshal(b)
	if err != nil {
		fmt.Println(err)
	}
	buildsBucket.Put(itob(b.ID), buildJSON)

	m.logger.Printf("Saving build %d for %s in %s", b.ID, c.Hash, p.ID)

	return tx.Commit()
}

func (m *Manager) QueueBuild(b *Build) error {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	buildQueueBucket := tx.Bucket([]byte("buildQueue"))
	if buildQueueBucket == nil {
		buildQueueBucket, err = tx.CreateBucket([]byte("buildQueue"))
		if err != nil {
			return err
		}
	}

	queuedBuild := NewQueuedBuild(b)
	queuedBuildJSON, err := json.Marshal(queuedBuild)
	if err != nil {
		fmt.Println(err)
	}

	now := uint64(time.Now().Unix())

	buildQueueBucket.Put(itob(now), queuedBuildJSON)

	return tx.Commit()
}

func (m *Manager) GetNextBuildID(p *project.Project, c *Commit) uint64 {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(p.ID))
	commitsBucket := projectBucket.Bucket([]byte("commits"))
	commitBucket := commitsBucket.Bucket([]byte(c.OrderedID()))

	buildsBucket, err := commitBucket.CreateBucketIfNotExists([]byte("builds"))
	if err != nil {
		log.Fatal(err)
	}
	if buildsBucket == nil {
		buildsBucket = commitBucket.Bucket([]byte("builds"))
	}

	id, _ := buildsBucket.NextSequence()
	tx.Commit()
	return id
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
