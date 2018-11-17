package builder_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
)

type BuilderSuite struct {
	suite.Suite
}

func TestBuilderSuite(t *testing.T) {
	suite.Run(t, new(BuilderSuite))
}

func (s *BuilderSuite) TestCreate() {
	m := builder.NewManager()

	b := m.Create()
	s.NotNil(b)

	s.NotEmpty(b.ID)
	s.NotEmpty(b.Token)
	s.Equal(builder.StateDisconnected, b.State)
	s.WithinDuration(time.Now().UTC(), b.CreatedAt, 1*time.Second)
	s.WithinDuration(time.Now().UTC(), b.UpdatedAt, 1*time.Second)
}

func (s *BuilderSuite) TestSave() {
	m := builder.NewManager()

	b := m.Create()
	s.NotNil(b)
	s.Equal(builder.StateDisconnected, b.State)
	b.State = builder.StateReady
	m.Save(b)

	b, err := m.GetByID(b.ID)
	s.Nil(err)
	s.Equal(builder.StateReady, b.State)
}

func (s *BuilderSuite) TestExists() {
	m := builder.NewManager()

	b := m.Create()

	s.True(m.Exists(b.ID))
}

func (s *BuilderSuite) TestDelete() {
	// validator, translator := domain.NewValidator()
	// m := builder.NewManager(s.storm, validator, translator, syncMock)

	// p, _ := m.Create("Test builder", velocity.GitRepository{
	// 	Address: "testGit",
	// })

	// err := m.Delete(p)
	// s.Nil(err)

	// s.False(m.Exists("Test builder"))
}
