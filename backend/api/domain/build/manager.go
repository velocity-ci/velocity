package build

import (
	"github.com/jinzhu/gorm"
)

type Manager struct {
	gormRepository *gormRepository
	fileManager    *fileManager
}

func NewManager(
	db *gorm.DB,
	fileManager *fileManager,
) *Manager {
	return &Manager{
		gormRepository: newGORMRepository(db),
		fileManager:    fileManager,
	}
}

func (m *Manager) SaveBuild(b Build) Build {
	return m.gormRepository.SaveBuild(b)
}

func (m *Manager) DeleteBuild(b Build) {
	m.gormRepository.DeleteBuild(b)
}

func (m *Manager) GetBuildByBuildID(id string) (Build, error) {
	return m.gormRepository.GetBuildByBuildID(id)
}

func (m *Manager) GetBuildsByProjectID(projectID string, q Query) ([]Build, uint64) {
	return m.gormRepository.GetBuildsByProjectID(projectID, q)
}

func (m *Manager) GetBuildsByCommitID(commitID string, q Query) ([]Build, uint64) {
	return m.gormRepository.GetBuildsByCommitID(commitID, q)
}

func (m *Manager) GetBuildsByTaskID(taskID string, q Query) ([]Build, uint64) {
	return m.gormRepository.GetBuildsByTaskID(taskID, q)
}

func (m *Manager) GetRunningBuilds() ([]Build, uint64) {
	return m.gormRepository.GetRunningBuilds()
}

func (m *Manager) GetWaitingBuilds() ([]Build, uint64) {
	return m.gormRepository.GetWaitingBuilds()
}

func (m *Manager) SaveBuildStep(bS BuildStep) BuildStep {
	return m.gormRepository.SaveBuildStep(bS)
}

func (m *Manager) DeleteBuildStep(bS BuildStep) {
	streams, _ := m.GetStreamsByBuildStepID(bS.ID)
	for _, bSS := range streams {
		m.DeleteStream(bSS)
	}
	m.gormRepository.DeleteBuildStep(bS)
}

func (m *Manager) GetBuildStepByBuildStepID(buildStepID string) (BuildStep, error) {
	return m.gormRepository.GetBuildStepByBuildStepID(buildStepID)
}
func (m *Manager) GetBuildStepsByBuildID(buildID string) ([]BuildStep, uint64) {
	return m.gormRepository.GetBuildStepsByBuildID(buildID)
}

func (m *Manager) SaveStream(s BuildStepStream) BuildStepStream {
	return m.gormRepository.SaveStream(s)
}

func (m *Manager) DeleteStream(s BuildStepStream) {
	m.gormRepository.DeleteStream(s)
	m.fileManager.DeleteByID(s.ID)
}

func (m *Manager) GetStreamsByBuildStepID(buildStepID string) ([]BuildStepStream, uint64) {
	return m.gormRepository.GetStreamsByBuildStepID(buildStepID)
}

func (m *Manager) GetStreamByID(id string) (BuildStepStream, error) {
	return m.gormRepository.GetStreamByID(id)
}

func (m *Manager) GetStreamByBuildStepIDAndStreamName(buildStepID string, name string) (BuildStepStream, error) {
	return m.gormRepository.GetStreamByBuildStepIDAndStreamName(buildStepID, name)
}

func (m *Manager) SaveStreamLine(sL StreamLine) StreamLine {
	return m.fileManager.SaveStreamLine(sL)
}

func (m *Manager) GetStreamLinesByStreamID(streamID string) ([]StreamLine, uint64) {
	return m.fileManager.GetByID(streamID)
}
