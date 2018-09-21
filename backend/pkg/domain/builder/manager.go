package builder

import (
	"fmt"
	"sync"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"

	uuid "github.com/satori/go.uuid"

	"github.com/velocity-ci/velocity/backend/pkg/domain/user"

	"github.com/velocity-ci/velocity/backend/pkg/domain"
)

// Event constants
const (
	EventCreate = "builder:new"
	EventUpdate = "builder:update"
	EventDelete = "builder:delete"
)

type Manager struct {
	builders map[string]*Builder
	mux      sync.Mutex

	brokers []domain.Broker
}

func NewManager() *Manager {
	return &Manager{
		brokers:  []domain.Broker{},
		builders: map[string]*Builder{},
	}
}

func (m *Manager) AddBroker(b domain.Broker) {
	m.brokers = append(m.brokers, b)
}

func (m *Manager) Create() *Builder {
	b := &Builder{
		ID:        uuid.NewV4().String(),
		Token:     user.GenerateRandomString(64),
		State:     StateDisconnected,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	m.Save(b)

	return b
}

func (m *Manager) Exists(id string) bool {
	if _, ok := m.builders[id]; ok {
		return true
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
		if v.State == StateReady {
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
		if v.State == StateBusy {
			r = append(r, v)
		}
	}

	return r, len(r) + skipCounter*q.Limit
}

func (m *Manager) Save(b *Builder) {
	m.mux.Lock()
	defer m.mux.Unlock()
	var ev string
	if m.Exists(b.ID) {
		ev = EventUpdate
	} else {
		ev = EventCreate
	}
	m.builders[b.ID] = b
	velocity.GetLogger().Debug("saved builder", zap.String("ID", b.ID))
	for _, br := range m.brokers {
		br.EmitAll(&domain.Emit{
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
	// if b.Command != nil && b.Command.Command == "build" {
	// 	build := b.Command.Payload.(*BuildCtrl).Build
	// 	build.Status = velocity.StateFailed
	// 	m.buildManager.Update(build)
	// }
	// delete(m.builders, b.ID)
}
