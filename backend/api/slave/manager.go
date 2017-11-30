package slave

import "github.com/velocity-ci/velocity/backend/api/domain/build"

type Manager struct {
	slaves       map[string]*Slave
	buildManager build.Repository
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

func (m *Manager) StartBuild(slave *Slave, b *build.Build) {
	// TODO: Sync known hosts
	slave.State = "busy"
	m.Save(slave)

	build.Status = "running"
	m.buildManager.SaveBuild(build)

	buildSteps, count := m.buildManager.GetBuildStepsForBuild(b)
	if count < 1 {
		for i, s := range b.Task.VTask.Steps {
			bS := build.NewBuildStep(
				b,
				i,
				s.GetDescription(),
			)
			m.buildManager.SaveBuildStep(bS)

			for _, oSName := range s.GetOutputStreams() {
				oS := build.NewOutputStream(bS, oSName)
				m.buildManager.SaveOutputStream(oS)
			}
			buildSteps = append(buildSteps, bS)
		}
	}
	slave.Command = NewBuildCommand(b, buildSteps)

	slave.ws.WriteJSON(slave.Command)
}
