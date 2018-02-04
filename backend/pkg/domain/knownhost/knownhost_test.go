package knownhost_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/asdine/storm"
	"github.com/stretchr/testify/suite"

	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"

	"github.com/velocity-ci/velocity/backend/pkg/domain"
)

type KnownHostSuite struct {
	suite.Suite
	kHM    *knownhost.Manager
	storm  *storm.DB
	dbPath string
}

func TestKnownHostSuite(t *testing.T) {
	suite.Run(t, new(KnownHostSuite))
}

func (s *KnownHostSuite) SetupTest() {
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
	s.kHM = knownhost.NewManager(s.storm, validator, translator)
}

func (s *KnownHostSuite) TearDownTest() {
	defer os.Remove(s.dbPath)
	s.storm.Close()
}

func (s *KnownHostSuite) TestValidNew() {
	m := s.kHM

	entry := `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`

	kH, errs := m.New(entry)
	s.Nil(errs)

	s.NotEmpty(kH.UUID)
	s.Equal(entry, kH.Entry)
	s.Equal([]string{"github.com"}, kH.Hosts)
	s.Equal("", kH.Comment)
	s.Equal("SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8", kH.SHA256Fingerprint)
	s.Equal("16:27:ac:a5:76:28:2d:36:63:1b:56:4d:eb:df:a6:48", kH.MD5Fingerprint)
}

func (s *KnownHostSuite) TestInvalidNew() {
	m := s.kHM

	entry := `github.com ssh-rsa AAAAB3NaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`

	kH, errs := m.New(entry)
	s.Nil(kH)
	s.NotNil(errs)

	s.Equal([]string{"entry is not a valid key!"}, errs.ErrorMap["entry"])
}

func (s *KnownHostSuite) TestSave() {
	m := s.kHM

	entry := `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`
	kH, _ := m.New(entry)

	err := m.Save(kH)

	s.Nil(err)

	s.True(m.Exists(entry))
}

func (s *KnownHostSuite) TestExists() {
	m := s.kHM

	entry := `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`
	s.False(m.Exists(entry))
}

func (s *KnownHostSuite) TestList() {
	m := s.kHM

	entry := `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`
	kH, _ := m.New(entry)
	m.Save(kH)

	q := &domain.PagingQuery{
		Limit: 3,
		Page:  1,
	}
	items, amount := m.GetAll(q)
	s.Len(items, 1)
	s.Equal(amount, 1)
}

func (s *KnownHostSuite) TestDelete() {
	m := s.kHM
	entry := `github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==`
	kH, _ := m.New(entry)
	m.Save(kH)

	err := m.Delete(kH)
	s.Nil(err)

	s.False(m.Exists(entry))
}
