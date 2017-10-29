package slave

import (
	"github.com/velocity-ci/velocity/backend/api/commit"
	"github.com/velocity-ci/velocity/backend/api/project"
)

type Manager struct {
	slaves map[string]*Slave
}

func NewManager() *Manager {
	return &Manager{
		slaves: map[string]*Slave{},
	}
}

func (m *Manager) Exists(slaveID string) bool {
	if _, ok := m.slaves[slaveID]; ok {
		return true
	}
	return false
}

func (m *Manager) WebSocketConnected(slaveID string) bool {
	if m.Exists(slaveID) {
		if m.slaves[slaveID].ws != nil {
			return true
		}
	}
	return false
}

func (m *Manager) GetSlaves() map[string]*Slave {
	return m.slaves
}

func (m *Manager) Save(s *Slave) {
	m.slaves[s.ID] = s
}

func (m *Manager) GetSlaveByID(slaveID string) *Slave {
	if m.Exists(slaveID) {
		return m.slaves[slaveID]
	}
	return nil
}

func (m *Manager) StartBuild(slaveID string, p *project.Project, commitHash string, build *commit.Build) {
	// TODO: Sync known hosts
	slave := m.GetSlaveByID(slaveID)
	slave.State = "busy"
	slave.Command = NewBuildCommand(p, build.Task, commitHash, build.ID)
	m.Save(slave)

	slave.ws.WriteJSON(slave.Command)
}
