package build

import (
	"time"

	"github.com/asdine/storm"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/velocity"
)

// Event constants
const (
	EventStepUpdate = "step:update"
)

type StepManager struct {
	db      *stepStormDB
	brokers []domain.Broker
}

func NewStepManager(
	db *storm.DB,
) *StepManager {
	m := &StepManager{
		db:      newStepStormDB(db),
		brokers: []domain.Broker{},
	}
	return m
}

func (m *StepManager) AddBroker(b domain.Broker) {
	m.brokers = append(m.brokers, b)
}

func (m *StepManager) create(
	b *Build,
	number int,
	vStep *velocity.Step,
) *Step {
	s := &Step{
		ID: uuid.NewV3(uuid.NewV1(), b.ID).String(),
		// Build:     b,
		Number:    number,
		VStep:     vStep,
		Status:    velocity.StateWaiting,
		UpdatedAt: time.Now().UTC(),
		Streams:   []*Stream{},
	}
	m.db.save(s)
	return s
}

func (m *StepManager) Update(s *Step) error {
	if err := m.db.save(s); err != nil {
		return err
	}
	for _, b := range m.brokers {
		b.EmitAll(&domain.Emit{
			Event:   EventStepUpdate,
			Payload: s,
		})
	}

	return nil
}

func (m *StepManager) GetByID(id string) (*Step, error) {
	return GetStepByID(m.db.DB, id)
}
