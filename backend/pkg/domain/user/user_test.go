package user_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/asdine/storm"
	"github.com/stretchr/testify/suite"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"
)

type UserSuite struct {
	suite.Suite
	uM     *user.Manager
	storm  *storm.DB
	dbPath string
}

func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}

func (s *UserSuite) SetupTest() {
	// Retrieve a temporary path.
	f, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	s.dbPath = f.Name()
	f.Close()
	os.Remove(s.dbPath)
	// Open the database.
	s.storm, err = storm.Open(s.dbPath)
	if err != nil {
		panic(err)
	}
	validator, translator := domain.NewValidator()
	s.uM = user.NewManager(s.storm, validator, translator)
}

func (s *UserSuite) TearDownTest() {
	defer os.Remove(s.dbPath)
	s.storm.Close()
}

func (s *UserSuite) TestValidNew() {
	m := s.uM
	u, errs := m.New("admin", "password")
	s.Nil(errs)

	s.NotEmpty(u.UUID)
	s.Equal("admin", u.Username)
	s.Empty(u.Password)
	s.NotEmpty(u.HashedPassword)
	s.True(u.ValidatePassword("password"))
}

func (s *UserSuite) TestInvalidNew() {
	m := s.uM

	u, errs := m.New("ad", "password")
	s.Nil(u)
	s.NotNil(errs)

	s.Equal([]string{"username must be at least 3 characters in length"}, errs.ErrorMap["username"])

	u, errs = m.New("admin", "pa")
	s.Nil(u)
	s.NotNil(errs)

	s.Equal([]string{"password must be at least 3 characters in length"}, errs.ErrorMap["password"])
}

func (s *UserSuite) TestSave() {
	m := s.uM

	u, _ := m.New("admin", "password")

	err := m.Save(u)

	s.Nil(err)

	s.True(m.Exists("admin"))
}

func (s *UserSuite) TestDelete() {
	m := s.uM

	u, _ := m.New("admin", "password")
	m.Save(u)

	err := m.Delete(u)
	s.Nil(err)

	s.False(m.Exists("admin"))
}
