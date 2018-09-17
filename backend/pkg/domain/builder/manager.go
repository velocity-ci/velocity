package builder

import (
	"fmt"
	"sync"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"

	"github.com/velocity-ci/velocity/backend/pkg/builder"
	"github.com/velocity-ci/velocity/backend/pkg/phoenix"

	uuid "github.com/satori/go.uuid"

	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"

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
	lock     sync.Mutex

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

func (m *Manager) CreateBuilder() *Builder {
	b := &Builder{
		ID:        uuid.NewV4().String(),
		Token:     user.GenerateRandomString(64),
		State:     stateReady,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	m.Save(b)

	return b
}

func (m *Manager) Exists(id string) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.builders[id]; ok {
		return true
	}
	return false
}

func (m *Manager) GetAll(q *domain.PagingQuery) (r []*Builder, t int) {
	m.lock.Lock()
	defer m.lock.Unlock()
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
	m.lock.Lock()
	defer m.lock.Unlock()
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
	m.lock.Lock()
	defer m.lock.Unlock()
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
	m.lock.Lock()
	velocity.GetLogger().Debug("saved builder", zap.String("ID", b.ID))
	var ev string
	if m.Exists(b.ID) {
		ev = EventUpdate
	} else {
		ev = EventCreate
	}
	m.builders[b.ID] = b
	m.lock.Unlock()
	for _, b := range m.brokers {
		b.EmitAll(&domain.Emit{
			Topic:   "builders",
			Event:   ev,
			Payload: b,
		})
	}
}

func (m *Manager) GetByID(id string) (*Builder, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.Exists(id) {
		return m.builders[id], nil
	}
	return nil, fmt.Errorf("could not find builder %s", id)
}

func (m *Manager) Delete(b *Builder) {
	// if b.Command != nil && b.Command.Command == "build" {
	// 	build := b.Command.Payload.(*BuildCtrl).Build
	// 	build.Status = velocity.StateFailed
	// 	m.buildManager.Update(build)
	// }
	// delete(m.builders, b.ID)
}

func (m *Manager) StartBuild(bu *Builder, b *build.Build) {
	bu.State = stateBusy
	m.Save(bu)

	// Set knownhosts
	knownHosts, _ := m.knownHostManager.GetAll(domain.NewPagingQuery())
	fmt.Printf("%+v\n", bu)
	fmt.Printf("%+v\n", bu.WS)
	fmt.Printf("%+v\n", bu.WS.Socket)
	resp := bu.WS.Socket.Send(&phoenix.PhoenixMessage{
		Event: builder.EventSetKnownHosts,
		Topic: fmt.Sprintf("builder:%s", bu.ID),
		Payload: &builder.KnownHostPayload{
			KnownHosts: knownHosts,
		},
	}, true)
	if resp.Status != phoenix.ResponseOK {
		velocity.GetLogger().Error("could not set knownhosts on builder", zap.String("builder", bu.ID))
	} else {
		velocity.GetLogger().Info("set knownhosts on builder", zap.String("builder", bu.ID))
	}

	// Start build
	b.Status = velocity.StateRunning
	m.buildManager.Update(b)

	steps := m.stepManager.GetStepsForBuild(b)
	streams := []*build.Stream{}
	for _, s := range steps {
		streams = append(streams, m.streamManager.GetStreamsForStep(s)...)
	}

	resp = bu.WS.Socket.Send(&phoenix.PhoenixMessage{
		Event: builder.EventStartBuild,
		Topic: fmt.Sprintf("builder:%s", bu.ID),
		Payload: &builder.BuildPayload{
			Build:   b,
			Steps:   steps,
			Streams: streams,
		},
	}, true)
	if resp.Status != phoenix.ResponseOK {
		velocity.GetLogger().Error("could not start build on builder", zap.String("builder", bu.ID))
	} else {
		velocity.GetLogger().Info("started build on builder", zap.String("builder", bu.ID))
	}
}
