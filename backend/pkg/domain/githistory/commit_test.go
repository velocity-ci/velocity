package githistory_test

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
)

func TestNew(t *testing.T) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	m := githistory.NewCommitManager(db)

	pM := project.NewManager(db, validator, translator, syncMock)
	p, _ := pM.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	ts := time.Now().UTC()

	c := m.New(p, "abcdef", "test commit", "me@velocityci.io", ts, []*githistory.Branch{})

	assert.NotNil(t, c)

	assert.Equal(t, p, c.Project)
	assert.Equal(t, "abcdef", c.Hash)
	assert.Equal(t, "test commit", c.Message)
	assert.Equal(t, "me@velocityci.io", c.Author)
	assert.Equal(t, ts, c.CreatedAt)
	assert.Empty(t, c.Branches)
}

func TestGetByProjectAndHash(t *testing.T) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	m := githistory.NewCommitManager(db)

	pM := project.NewManager(db, validator, translator, syncMock)
	p, _ := pM.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	ts := time.Now()

	nC := m.New(p, "abcdef", "test commit", "me@velocityci.io", ts, []*githistory.Branch{})

	m.Save(nC)

	c, err := m.GetByProjectAndHash(p, nC.Hash)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	assert.EqualValues(t, nC, c)
}
