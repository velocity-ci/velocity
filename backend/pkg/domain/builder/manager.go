package builder

import (
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"

	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"

	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
)

// Event constants
const (
	EventCreate = "builder:new"
	EventUpdate = "builder:update"
	EventDelete = "builder:delete"
)

type Manager struct {
	builders map[string]*Builder

	brokers []domain.Broker

	buildManager     *build.BuildManager
	stepManager      *build.StepManager
	streamManager    *build.StreamManager
	knownHostManager *knownhost.Manager
}

func NewManager(
	buildManager *build.BuildManager,
	knownhostManager *knownhost.Manager,
	stepManager *build.StepManager,
	streamManager *build.StreamManager,
) *Manager {
	return &Manager{
		buildManager:     buildManager,
		knownHostManager: knownhostManager,
		brokers:          []domain.Broker{},
		stepManager:      stepManager,
		streamManager:    streamManager,
		builders:         map[string]*Builder{},
	}
}

func (m *Manager) AddBroker(b domain.Broker) {
	m.brokers = append(m.brokers, b)
}

func (m *Manager) CreateBuilder(t Transport) *Builder {
	b := &Builder{
		ID:        uuid.NewV4().String(),
		State:     stateReady,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),

		ws: t,
	}
	m.Save(b)

	go m.monitor(b)

	return b
}

func (m *Manager) Exists(id string) bool {
	if _, ok := m.builders[id]; ok {
		return true
	}
	return false
}

func (m *Manager) WebsocketConnected(id string) bool {
	if m.Exists(id) {
		if m.builders[id].ws != nil {
			return true
		}
	}
	return false
}

func (m *Manager) GetAll(q *domain.PagingQuery) (r []*Builder, t int) {
	t = 0

	skipCounter := 0
	for _, v := range m.builders {
		if len(r) >= q.Limit {
			break
		}
		if skipCounter < (q.Page-1)*q.Limit {
			skipCounter++
			break
		}
		r = append(r, v)
	}

	return r, len(r) + skipCounter*q.Limit
}

func (m *Manager) GetReady(q *domain.PagingQuery) (r []*Builder, t int) {
	t = 0

	skipCounter := 0
	for _, v := range m.builders {
		if len(r) >= q.Limit {
			break
		}
		if skipCounter < (q.Page-1)*q.Limit {
			skipCounter++
			break
		}
		if v.State == stateReady {
			r = append(r, v)
		}
	}

	return r, len(r) + skipCounter*q.Limit
}

func (m *Manager) GetBusy(q *domain.PagingQuery) (r []*Builder, t int) {
	t = 0

	skipCounter := 0
	for _, v := range m.builders {
		if len(r) >= q.Limit {
			break
		}
		if skipCounter < (q.Page-1)*q.Limit {
			skipCounter++
			break
		}
		if v.State == stateBusy {
			r = append(r, v)
		}
	}

	return r, len(r) + skipCounter*q.Limit
}

func (m *Manager) Save(b *Builder) {
	var ev string
	if m.Exists(b.ID) {
		ev = EventUpdate
	} else {
		ev = EventCreate
	}
	m.builders[b.ID] = b
	for _, b := range m.brokers {
		b.EmitAll(&domain.Emit{
			Topic:   "builders",
			Event:   ev,
			Payload: b,
		})
	}
}

func (m *Manager) GetByID(id string) (*Builder, error) {
	if m.Exists(id) {
		return m.builders[id], nil
	}
	return nil, fmt.Errorf("could not find builder %s", id)
}

func (m *Manager) Delete(b *Builder) {
	if b.Command != nil && b.Command.Command == "build" {
		build := b.Command.Payload.(*BuildCtrl).Build
		build.Status = velocity.StateFailed
		m.buildManager.Update(build)
	}
	delete(m.builders, b.ID)
}

func (m *Manager) StartBuild(builder *Builder, b *build.Build) {
	builder.State = stateBusy
	m.Save(builder)

	// Add knownhosts
	knownHosts, _ := m.knownHostManager.GetAll(domain.NewPagingQuery())
	builder.Command = newKnownHostsCommand(knownHosts)
	builder.ws.WriteJSON(builder.Command)

	// Start build
	b.Status = velocity.StateRunning
	m.buildManager.Update(b)

	steps := m.stepManager.GetStepsForBuild(b)
	streams := []*build.Stream{}
	for _, s := range steps {
		streams = append(streams, m.streamManager.GetStreamsForStep(s)...)
	}

	builder.Command = newBuildCommand(b, steps, streams)
	builder.ws.WriteJSON(builder.Command)
}
