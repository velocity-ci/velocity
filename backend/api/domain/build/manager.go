package build

import (
	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type Manager struct {
	gormRepository *gormRepository
	fileManager    *fileManager
}

func NewManager(
	db *gorm.DB,
) *Manager {
	return &Manager{
		gormRepository: newGORMRepository(db),
	}
}

func (m *Manager) SaveBuild(b Build) Build {
	return m.gormRepository.SaveBuild(b)
}
func (m *Manager) DeleteBuild(b Build) {
	m.gormRepository.DeleteBuild(b)
}
func (m *Manager) GetBuildByProjectAndCommitAndID(p project.Project, c commit.Commit, id string) (Build, error) {
	return m.gormRepository.GetBuildByProjectAndCommitAndID(p, c, id)
}
func (m *Manager) GetBuildsByProject(p project.Project, q Query) ([]Build, uint64) {
	return m.gormRepository.GetBuildsByProject(p, q)
}
func (m *Manager) GetBuildsByProjectAndCommit(p project.Project, c commit.Commit) ([]Build, uint64) {
	return m.gormRepository.GetBuildsByProjectAndCommit(p, c)
}
func (m *Manager) SaveBuildStep(bS BuildStep) BuildStep {
	return m.gormRepository.SaveBuildStep(bS)
}
func (m *Manager) GetBuildStepsForBuild(b Build) ([]BuildStep, uint64) {
	return m.gormRepository.GetBuildStepsForBuild(b)
}
func (m *Manager) GetBuildStepByBuildAndID(b Build, id string) (BuildStep, error) {
	return m.gormRepository.GetBuildStepByBuildAndID(b, id)
}
func (m *Manager) SaveOutputStream(oS velocity.OutputStream) velocity.OutputStream {
	// return m.gormRepository.SaveOutputStream(oS)
	return oS
}
func (m *Manager) GetOutputStreamsForBuildStep(bS BuildStep) ([]velocity.OutputStream, uint64) {
	return m.gormRepository.GetOutputStreamsForBuildStep(bS)
}

func (m *Manager) GetRunningBuilds() ([]Build, uint64) {
	return m.gormRepository.GetRunningBuilds()
}
func (m *Manager) GetWaitingBuilds() ([]Build, uint64) {
	return m.gormRepository.GetWaitingBuilds()
}

func (m *Manager) GetOutputStreamByID(id string) (velocity.OutputStream, error) {
	return m.gormRepository.GetOutputStreamByID(id)
}

func (m *Manager) SaveStreamLine(sL StreamLine) StreamLine {
	return m.fileManager.SaveStreamLine(sL)
}
