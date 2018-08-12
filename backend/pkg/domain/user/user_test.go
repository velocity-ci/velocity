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

func (s *UserSuite) TestValidCreate() {
	m := s.uM
	u, errs := m.Create("admin", "password")
	s.Nil(errs)

	s.NotEmpty(u.ID)
	s.Equal("admin", u.Username)
	s.Empty(u.Password)
	s.NotEmpty(u.HashedPassword)
	s.True(u.ValidatePassword("password"))
}

func (s *UserSuite) TestInvalidNew() {
	m := s.uM

	u, errs := m.Create("ad", "password")
	s.Nil(u)
	s.NotNil(errs)

	s.Equal([]string{"username must be at least 3 characters in length"}, errs.ErrorMap["username"])

	u, errs = m.Create("admin", "pa")
	s.Nil(u)
	s.NotNil(errs)

	s.Equal([]string{"password must be at least 3 characters in length"}, errs.ErrorMap["password"])
}

func (s *UserSuite) TestUnique() {
	m := s.uM

	u, errs := m.Create("user", "password")
	s.NotNil(u)
	s.Nil(errs)

	u, errs = m.Create("user", "pass")
	s.Nil(u)
	s.NotNil(errs)

	s.Equal([]string{"username already exists!"}, errs.ErrorMap["username"])
}

func (s *UserSuite) TestSave() {
	m := s.uM

	u, err := m.Create("admin", "password")
	s.Nil(err)
	s.NotNil(u)

	s.True(m.Exists("admin"))
}

func (s *UserSuite) TestDelete() {
	m := s.uM

	u, _ := m.Create("admin", "password")

	err := m.Delete(u)
	s.Nil(err)

	s.False(m.Exists("admin"))
}

func (s *UserSuite) TestEnsureAdminIfNoAdmin() {
	s.uM.EnsureAdminUser()

	u, err := s.uM.GetByUsername("admin")
	s.Nil(err)
	s.NotNil(u)
}

func (s *UserSuite) TestEnsureAdminIfAlreadyAdmin() {
	eU, _ := s.uM.Create("admin", "password1234")

	s.uM.EnsureAdminUser()

	u, _ := s.uM.GetByUsername("admin")

	s.Equal(eU, u)
}
