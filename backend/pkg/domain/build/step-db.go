package build

import (
	"encoding/json"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/labstack/gommon/log"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type GormStep struct {
	UUID string `gorm:"primary_key"`
	// Build       *GormBuild `gorm:"ForeignKey:BuildID"`
	// BuildID     string
	Number      int
	Status      string
	VStep       []byte
	Streams     []*GormStream
	StartedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt time.Time
}

func (GormStep) TableName() string {
	return "build_steps"
}

func (g *GormStep) ToStep() *Step {
	var gStep map[string]interface{}
	err := json.Unmarshal(g.VStep, &gStep)
	if err != nil {
		logrus.Error(err)
	}

	vStep, err := velocity.DetermineStepFromInterface(gStep)
	if err != nil {
		logrus.Error(err)
	} else {
		json.Unmarshal(g.VStep, vStep)
	}

	streams := []*Stream{}
	for _, s := range g.Streams {
		streams = append(streams, s.ToStream())
	}

	return &Step{
		UUID: g.UUID,
		// Build:       g.Build.ToBuild(),
		Number:      g.Number,
		Status:      g.Status,
		VStep:       &vStep,
		Streams:     streams,
		StartedAt:   g.StartedAt,
		UpdatedAt:   g.UpdatedAt,
		CompletedAt: g.CompletedAt,
	}
}

func (s *Step) ToGormStep() *GormStep {
	jsonStep, err := json.Marshal(s.VStep)
	if err != nil {
		log.Error(err)
	}
	streams := []*GormStream{}
	for _, s := range s.Streams {
		streams = append(streams, s.ToGormStream())
	}

	return &GormStep{
		UUID: s.UUID,
		// Build:       s.Build.ToGormBuild(),
		Number:      s.Number,
		VStep:       jsonStep,
		Streams:     streams,
		StartedAt:   s.StartedAt,
		UpdatedAt:   s.UpdatedAt,
		CompletedAt: s.CompletedAt,
	}
}

type stepDB struct {
	db *gorm.DB
}

func newStepDB(gorm *gorm.DB) *stepDB {
	return &stepDB{
		db: gorm,
	}
}

func (db *stepDB) save(s *Step) error {
	tx := db.db.Begin()

	g := s.ToGormStep()

	tx.
		Where(GormStep{UUID: s.UUID}).
		Assign(&g).
		FirstOrCreate(&g)

	return tx.Commit().Error
}
