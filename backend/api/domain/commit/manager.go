package commit

import (
	"github.com/jinzhu/gorm"
)

type Manager struct {
	gormRepository *gormRepository
}

func NewManager(
	db *gorm.DB,
) *Manager {
	return &Manager{
		gormRepository: newGORMRepository(db),
	}
}

func (m *Manager) SaveCommit(c Commit) Commit {
	m.gormRepository.SaveCommit(c)
	return c
}

func (m *Manager) DeleteCommit(c Commit) {
	m.gormRepository.DeleteCommit(c)
}

func (m *Manager) GetCommitByCommitID(id string) (Commit, error) {
	return m.gormRepository.GetCommitByCommitID(id)
}

func (m *Manager) GetCommitByProjectIDAndCommitHash(projectID string, hash string) (Commit, error) {
	return m.gormRepository.GetCommitByProjectIDAndCommitHash(projectID, hash)
}

func (m *Manager) GetAllCommitsByProjectID(projectID string, q Query) ([]Commit, uint64) {
	return m.gormRepository.GetAllCommitsByProjectID(projectID, q)
}

func (m *Manager) SaveBranch(b Branch) Branch {
	return m.gormRepository.SaveBranch(b)
}
func (m *Manager) DeleteBranch(b Branch) {
	m.gormRepository.DeleteBranch(b)
}
func (m *Manager) GetBranchByProjectIDAndName(projectID string, name string) (Branch, error) {
	return m.gormRepository.GetBranchByProjectIDAndName(projectID, name)
}
func (m *Manager) GetAllBranchesByProjectID(projectID string, q Query) ([]Branch, uint64) {
	return m.gormRepository.GetAllBranchesByProjectID(projectID, q)
}
