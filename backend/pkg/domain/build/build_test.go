package build_test

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
)

func TestNewBuild(t *testing.T) {
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

	tM := task.NewManager(db)
	tsk := tM.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(db, build.NewStepManager(db), build.NewStreamManager(db))
	params := map[string]string{}
	b := m.New(tsk, params)

	assert.Equal(t, tsk, b.Task)
	assert.Equal(t, params, b.Parameters)
	assert.Equal(t, velocity.StateWaiting, b.Status)
	assert.WithinDuration(t, time.Now().UTC(), b.CreatedAt, 1*time.Second)
	assert.WithinDuration(t, time.Now().UTC(), b.UpdatedAt, 1*time.Second)

	assert.Len(t, b.Steps, len(tsk.Steps))
}

func TestSaveBuild(t *testing.T) {
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

	tM := task.NewManager(db)
	tsk := tM.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(db, build.NewStepManager(db), build.NewStreamManager(db))
	params := map[string]string{}
	b := m.New(tsk, params)

	err := m.Save(b)
	assert.Nil(t, err)
}

func TestGetBuildsForProject(t *testing.T) {
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

	tM := task.NewManager(db)
	tsk := tM.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(db, build.NewStepManager(db), build.NewStreamManager(db))
	params := map[string]string{}
	b := m.New(tsk, params)
	m.Save(b)

	rbs, total := m.GetAllForProject(p, &domain.PagingQuery{Limit: 5, Page: 1})

	assert.Equal(t, 1, total)
	assert.Len(t, rbs, 1)

	assert.Equal(t, b, rbs[0])
}

func TestGetBuildsForCommit(t *testing.T) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	pM := project.NewManager(db, validator, translator, syncMock)
	cM := githistory.NewCommitManager(db)
	bM := githistory.NewBranchManager(db)

	p, _ := pM.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := bM.New(p, "testBranch")

	c := cM.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), []*githistory.Branch{br})

	tM := task.NewManager(db)
	tsk := tM.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(db, build.NewStepManager(db), build.NewStreamManager(db))
	params := map[string]string{}
	b := m.New(tsk, params)
	m.Save(b)

	rbs, total := m.GetAllForCommit(c, &domain.PagingQuery{Limit: 5, Page: 1})

	assert.Equal(t, 1, total)
	assert.Len(t, rbs, 1)

	assert.Equal(t, b, rbs[0])
}

func TestGetBuildsForBranch(t *testing.T) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	pM := project.NewManager(db, validator, translator, syncMock)
	cM := githistory.NewCommitManager(db)
	bM := githistory.NewBranchManager(db)

	p, _ := pM.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := bM.New(p, "testBranch")

	c := cM.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), []*githistory.Branch{br})

	tM := task.NewManager(db)
	tsk := tM.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(db, build.NewStepManager(db), build.NewStreamManager(db))
	params := map[string]string{}
	b := m.New(tsk, params)
	m.Save(b)

	rbs, total := m.GetAllForBranch(br, &domain.PagingQuery{Limit: 5, Page: 1})

	assert.Equal(t, 1, total)
	assert.Len(t, rbs, 1)

	assert.Equal(t, rbs[0], b)
}

func TestGetBuildsForTask(t *testing.T) {
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

	tM := task.NewManager(db)
	tsk := tM.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(db, build.NewStepManager(db), build.NewStreamManager(db))
	params := map[string]string{}
	b := m.New(tsk, params)
	m.Save(b)

	rbs, total := m.GetAllForTask(tsk, &domain.PagingQuery{Limit: 5, Page: 1})

	assert.Equal(t, 1, total)
	assert.Len(t, rbs, 1)

	assert.Equal(t, b, rbs[0])
}

func TestGetRunningBuilds(t *testing.T) {
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

	tM := task.NewManager(db)
	tsk := tM.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(db, build.NewStepManager(db), build.NewStreamManager(db))
	params := map[string]string{}
	b := m.New(tsk, params)
	b.Status = velocity.StateRunning
	m.Save(b)

	rbs, total := m.GetRunningBuilds()

	assert.Equal(t, 1, total)
	assert.Len(t, rbs, 1)

	assert.Equal(t, b, rbs[0])
}

func TestGetWaitingBuilds(t *testing.T) {
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

	tM := task.NewManager(db)
	tsk := tM.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(db, build.NewStepManager(db), build.NewStreamManager(db))
	params := map[string]string{}
	b := m.New(tsk, params)
	m.Save(b)

	rbs, total := m.GetWaitingBuilds()

	assert.Equal(t, 1, total)
	assert.Len(t, rbs, 1)

	assert.Equal(t, b, rbs[0])
}

func TestGetBuildByUUID(t *testing.T) {
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

	tM := task.NewManager(db)
	tsk := tM.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(db, build.NewStepManager(db), build.NewStreamManager(db))
	params := map[string]string{}
	b := m.New(tsk, params)
	m.Save(b)

	rB, err := m.GetBuildByUUID(b.UUID)
	assert.Nil(t, err)
	assert.Equal(t, b, rB)
}
