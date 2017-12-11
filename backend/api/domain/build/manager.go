package build

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/api/websocket"
)

type Manager struct {
	gormRepository   *gormRepository
	fileManager      *fileManager
	websocketManager *websocket.Manager
}

func NewManager(
	db *gorm.DB,
	fileManager *fileManager,
	websocketManager *websocket.Manager,
) *Manager {
	return &Manager{
		gormRepository:   newGORMRepository(db),
		fileManager:      fileManager,
		websocketManager: websocketManager,
	}
}

func (m *Manager) CreateBuild(b Build) Build {
	b.CreatedAt = time.Now()
	m.gormRepository.SaveBuild(b)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   fmt.Sprintf("project:%s", b.ProjectID),
		Event:   websocket.VNewBuild,
		Payload: NewResponseBuild(b),
	})
	return b
}

func (m *Manager) UpdateBuild(b Build) Build {
	m.gormRepository.SaveBuild(b)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   fmt.Sprintf("project:%s", b.ProjectID),
		Event:   websocket.VUpdateBuild,
		Payload: NewResponseBuild(b),
	})
	return b
}

func (m *Manager) DeleteBuild(b Build) {
	m.gormRepository.DeleteBuild(b)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   fmt.Sprintf("project:%s", b.ProjectID),
		Event:   websocket.VDeleteBuild,
		Payload: NewResponseBuild(b),
	})
}

func (m *Manager) GetBuildByBuildID(id string) (Build, error) {
	return m.gormRepository.GetBuildByBuildID(id)
}

func (m *Manager) GetBuildsByProjectID(projectID string, q BuildQuery) ([]Build, uint64) {
	return m.gormRepository.GetBuildsByProjectID(projectID, q)
}

func (m *Manager) GetBuildsByCommitID(commitID string, q BuildQuery) ([]Build, uint64) {
	return m.gormRepository.GetBuildsByCommitID(commitID, q)
}

func (m *Manager) GetBuildsByTaskID(taskID string, q BuildQuery) ([]Build, uint64) {
	return m.gormRepository.GetBuildsByTaskID(taskID, q)
}

func (m *Manager) GetRunningBuilds() ([]Build, uint64) {
	return m.gormRepository.GetRunningBuilds()
}

func (m *Manager) GetWaitingBuilds() ([]Build, uint64) {
	return m.gormRepository.GetWaitingBuilds()
}

func (m *Manager) CreateBuildStep(bS BuildStep) BuildStep {
	return m.gormRepository.SaveBuildStep(bS)
}

func (m *Manager) UpdateBuildStep(bS BuildStep) BuildStep {
	m.gormRepository.SaveBuildStep(bS)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   fmt.Sprintf("step:%s", bS.ID),
		Event:   websocket.VUpdateStep,
		Payload: NewWebsocketBuildStep(bS),
	})
	return bS
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

func (m *Manager) GetBuildStepByBuildIDAndNumber(buildID string, stepNumber uint64) (BuildStep, error) {
	return m.gormRepository.GetBuildStepByBuildIDAndNumber(buildID, stepNumber)
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

func (m *Manager) CreateStreamLine(sL StreamLine) StreamLine {
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   fmt.Sprintf("stream:%s", sL.BuildStepStreamID),
		Event:   websocket.VNewStreamLine,
		Payload: sL,
	})
	m.fileManager.SaveStreamLine(sL)
	return sL
}

func (m *Manager) GetStreamLinesByStreamID(streamID string, q StreamLineQuery) ([]StreamLine, uint64) {
	return m.fileManager.GetByID(streamID, q)
}
