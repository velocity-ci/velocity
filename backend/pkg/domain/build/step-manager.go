package build

import (
	"time"

	"github.com/asdine/storm"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type StepManager struct {
	db *stepStormDB
}

func NewStepManager(
	db *storm.DB,
) *StepManager {
	m := &StepManager{
		db: newStepStormDB(db),
	}
	return m
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
	return m.db.save(s)
}

func (m *StepManager) GetByID(id string) (*Step, error) {
	return GetStepByID(m.db.DB, id)
}
