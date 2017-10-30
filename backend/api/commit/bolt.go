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
	"github.com/velocity-ci/velocity/backend/velocity"
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

func (m *Manager) GetCommitInProject(hash string, projectID string) (*Commit, error) {
	tx, err := m.bolt.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(projectID))
	commitsBucket := projectBucket.Bucket([]byte("commits"))
	if commitsBucket == nil {
		return nil, fmt.Errorf("Could not find any commits for project: %s", projectID)
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

	return nil, fmt.Errorf("Could not find project: %s, commit: %s", projectID, hash)
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

func (m *Manager) SaveTaskForCommitInProject(t *velocity.Task, c *Commit, p *project.Project) error {
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

func (m *Manager) GetTasksForCommitInProject(c *Commit, p *project.Project) []velocity.Task {
	tasks := []velocity.Task{}

	tx, err := m.bolt.Begin(false)
	if err != nil {
		log.Fatal(err)
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
		task := velocity.NewTask()
		err := json.Unmarshal(v, &task)
		if err == nil {
			tasks = append(tasks, task)
		} else {
			log.Fatal(err)
		}
	}

	return tasks
}

func (m *Manager) GetTaskForCommitInProject(c *Commit, p *project.Project, name string) (*velocity.Task, error) {

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

	task := velocity.NewTask()
	err = json.Unmarshal(t, &task)

	if err != nil {
		return nil, fmt.Errorf("Could not find commit for project: %s", p.ID)
	}

	return &task, nil
}

func (m *Manager) GetTotalCommitsForProject(p *project.Project, queryOpts *CommitQueryOpts) uint {
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
				count++
			}
		}
	}

	return count
}

func (m *Manager) SaveBuild(b *Build, projectID string, commitHash string) error {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(projectID))
	commitsBucket := projectBucket.Bucket([]byte("commits"))
	commit, _ := m.GetCommitInProject(commitHash, projectID)
	commitBucket := commitsBucket.Bucket([]byte(commit.OrderedID()))

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

	m.logger.Printf("Saving build %d for %s in %s", b.ID, commitHash, projectID)

	return tx.Commit()
}

func (m *Manager) QueueBuild(b *QueuedBuild) error {
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

	queuedBuildJSON, err := json.Marshal(b)
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
	return id
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func (m *Manager) GetBuildFromQueuedBuild(queuedBuild *QueuedBuild) *Build {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(queuedBuild.ProjectID))
	commitsBucket := projectBucket.Bucket([]byte("commits"))
	commit, _ := m.GetCommitInProject(queuedBuild.CommitHash, queuedBuild.ProjectID)
	commitBucket := commitsBucket.Bucket([]byte(commit.OrderedID()))

	buildsBucket, err := commitBucket.CreateBucketIfNotExists([]byte("builds"))
	if err != nil {
		log.Fatal(err)
	}
	if buildsBucket == nil {
		buildsBucket = commitBucket.Bucket([]byte("builds"))
	}

	buildJSON := buildsBucket.Get(itob(queuedBuild.ID))

	build := &Build{}
	err = json.Unmarshal(buildJSON, build)

	return build
}

func (m *Manager) GetBuild(projectID string, commitHash string, buildID uint64) *Build {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	projectBucket := projectsBucket.Bucket([]byte(projectID))
	commitsBucket := projectBucket.Bucket([]byte("commits"))
	commit, _ := m.GetCommitInProject(commitHash, projectID)
	commitBucket := commitsBucket.Bucket([]byte(commit.OrderedID()))

	buildsBucket := commitBucket.Bucket([]byte("builds"))
	if buildsBucket == nil {
		return nil
	}

	buildJSON := buildsBucket.Get(itob(buildID))

	build := &Build{}
	err = json.Unmarshal(buildJSON, build)

	return build
}

func (m *Manager) GetQueuedBuilds() []*QueuedBuild {

	queuedBuilds := []*QueuedBuild{}

	tx, err := m.bolt.Begin(true)
	if err != nil {
		return queuedBuilds
	}
	defer tx.Rollback()

	buildQueueBucket := tx.Bucket([]byte("buildQueue"))
	if buildQueueBucket == nil {
		return queuedBuilds
	}

	cursor := buildQueueBucket.Cursor()
	for k, v := cursor.Last(); k != nil; k, v = cursor.Prev() {
		queuedBuild := &QueuedBuild{}
		err := json.Unmarshal(v, queuedBuild)
		if err == nil {
			queuedBuilds = append(queuedBuilds, queuedBuild)
		}
	}

	return queuedBuilds
}

func (m *Manager) RemoveQueuedBuild(projectID string, commitHash string, buildID uint64) {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return
	}
	defer tx.Rollback()

	buildQueueBucket := tx.Bucket([]byte("buildQueue"))
	if buildQueueBucket == nil {
		return
	}

	cursor := buildQueueBucket.Cursor()
	for k, v := cursor.Last(); k != nil; k, v = cursor.Prev() {
		queuedBuild := &QueuedBuild{}
		err := json.Unmarshal(v, queuedBuild)
		if err == nil {
			if queuedBuild.ProjectID == projectID &&
				queuedBuild.CommitHash == commitHash &&
				queuedBuild.ID == buildID {
				buildQueueBucket.Delete(k)
			}
		}
	}

	tx.Commit()
}
