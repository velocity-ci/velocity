package git

import "os"

import "io/ioutil"

// RepositoryManager manages repositories on the local host
type RepositoryManager struct {
	localRepositories map[string]*RawRepository
}

// NewRepositoryManager returns a new repository manager
func NewRepositoryManager() *RepositoryManager {
	return &RepositoryManager{
		localRepositories: map[string]*RawRepository{},
	}
}

// Add adds a repository to be managed on the local host
func (m *RepositoryManager) Add(address string, privateKey string, hostKey string) error {

	dir, err := ioutil.TempDir("", "velocity-repository")
	if err != nil {
		return err
	}

	rawRepo, err := Clone(&Repository{
		Address:    address,
		PrivateKey: privateKey,
	}, &CloneOptions{}, dir, os.Stdout)
	if err != nil {
		return err
	}
	m.localRepositories[address] = rawRepo

	return nil
}
