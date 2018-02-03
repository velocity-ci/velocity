package build

import (
	"time"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type StepManager struct {
	db *stepDB
}

func NewStepManager(
	db *gorm.DB,
) *StepManager {
	db.AutoMigrate(&GormBuild{}, &GormStep{}, &GormStream{})
	m := &StepManager{
		db: newStepDB(db),
	}
	return m
}

func (m *StepManager) new(
	b *Build,
	number int,
	vStep *velocity.Step,
) *Step {
	return &Step{
		UUID: uuid.NewV3(uuid.NewV1(), b.UUID).String(),
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
