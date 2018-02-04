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

func (m *StepManager) new(
	b *Build,
	number int,
	vStep *velocity.Step,
) *Step {
	return &Step{
		ID: uuid.NewV3(uuid.NewV1(), b.ID).String(),
		// Build:     b,
		Number:    number,
		VStep:     vStep,
		Status:    velocity.StateWaiting,
		UpdatedAt: time.Now().UTC(),
		Streams:   []*Stream{},
	}
}

func (m *StepManager) Save(s *Step) error {
	return m.db.save(s)
}
