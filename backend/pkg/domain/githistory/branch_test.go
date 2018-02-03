package githistory_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
)

func TestNewBranch(t *testing.T) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	m := githistory.NewBranchManager(db)

	pM := project.NewManager(db, validator, translator, syncMock)
	p, _ := pM.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	b := m.New(p, "testBranch")
	assert.NotNil(t, b)

	assert.Equal(t, p, b.Project)
	assert.Equal(t, "testBranch", b.Name)
}

func TestSaveBranch(t *testing.T) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	m := githistory.NewBranchManager(db)

	pM := project.NewManager(db, validator, translator, syncMock)
	p, _ := pM.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	b := m.New(p, "testBranch")

	err := m.Save(b)

	assert.Nil(t, err)
}

func TestGetAllBranchesForProject(t *testing.T) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	m := githistory.NewBranchManager(db)

	pM := project.NewManager(db, validator, translator, syncMock)
	p, _ := pM.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	b1 := m.New(p, "testBranch")
	b2 := m.New(p, "2estBranch")

	m.Save(b1)
	m.Save(b2)

	bs, total := m.GetAllForProject(p, &domain.PagingQuery{Limit: 5, Page: 1})

	assert.Equal(t, 2, total)
	assert.Len(t, bs, 2)
	assert.Equal(t, bs[0], b1)
	assert.Equal(t, bs[1], b2)
}
