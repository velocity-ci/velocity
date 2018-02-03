package task_test

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
)

func TestNew(t *testing.T) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	cM := githistory.NewCommitManager(db)

	pM := project.NewManager(db, validator, translator, syncMock)
	p, _ := pM.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	c := cM.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), []*githistory.Branch{})

	m := task.NewManager(db)
	tsk := m.New(c, &velocity.Task{
		Name: "testTask",
	})

	assert.NotNil(t, tsk)

	assert.Equal(t, c, tsk.Commit)
	assert.Equal(t, "testTask", tsk.Name)
}

func TestSave(t *testing.T) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	cM := githistory.NewCommitManager(db)

	pM := project.NewManager(db, validator, translator, syncMock)
	p, _ := pM.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	c := cM.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), []*githistory.Branch{})

	m := task.NewManager(db)
	tsk := m.New(c, &velocity.Task{
		Name: "testTask",
	})

	err := m.Save(tsk)
	assert.Nil(t, err)
}

func TestGetByCommitAndName(t *testing.T) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	cM := githistory.NewCommitManager(db)

	pM := project.NewManager(db, validator, translator, syncMock)
	p, _ := pM.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	c := cM.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), []*githistory.Branch{})

	m := task.NewManager(db)
	tsk := m.New(c, &velocity.Task{
		Name: "testTask",
	})

	m.Save(tsk)

	rTsk, err := m.GetByCommitAndName(c, "testTask")
	assert.NotNil(t, rTsk)
	assert.Nil(t, err)

	assert.Equal(t, tsk, rTsk)
}

func TestGetAllForCommit(t *testing.T) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	cM := githistory.NewCommitManager(db)

	pM := project.NewManager(db, validator, translator, syncMock)
	p, _ := pM.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	c := cM.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), []*githistory.Branch{})

	m := task.NewManager(db)
	tsk1 := m.New(c, &velocity.Task{
		Name: "testTask",
	})
	tsk2 := m.New(c, &velocity.Task{
		Name: "2estTask",
	})

	m.Save(tsk1)
	m.Save(tsk2)

	rTsks, total := m.GetAllForCommit(c, &domain.PagingQuery{Limit: 5, Page: 1})

	assert.Equal(t, 2, total)
	assert.Len(t, rTsks, 2)

	assert.Equal(t, rTsks[0], tsk1)
	assert.Equal(t, rTsks[1], tsk2)
}
