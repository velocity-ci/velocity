package build

import (
	"encoding/json"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type stormStep struct {
	ID      string `storm:"id"`
	Number  int
	VStep   []byte
	Streams []stormStream

	Status      string
	UpdatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
}

func (s *stormStep) toStep() *Step {
	var gStep map[string]interface{}
	err := json.Unmarshal(s.VStep, &gStep)
	if err != nil {
		logrus.Error(err)
	}

	vStep, err := velocity.DetermineStepFromInterface(gStep)
	if err != nil {
		logrus.Error(err)
	} else {
		json.Unmarshal(s.VStep, vStep)
	}

	streams := []*Stream{}
	for _, s := range s.Streams {
		streams = append(streams, s.toStream())
	}

	return &Step{
		ID:          s.ID,
		Number:      s.Number,
		Status:      s.Status,
		UpdatedAt:   s.UpdatedAt,
		StartedAt:   s.StartedAt,
		CompletedAt: s.CompletedAt,
		VStep:       &vStep,
		Streams:     streams,
	}
}

func (s *Step) toStormStep() stormStep {
	stepJSON, err := json.Marshal(s.VStep)
	if err != nil {
		logrus.Error(err)
	}
	streams := []stormStream{}
	for _, s := range s.Streams {
		streams = append(streams, s.toStormStream())
	}
	return stormStep{
		ID:          s.ID,
		Number:      s.Number,
		VStep:       stepJSON,
		Streams:     streams,
		Status:      s.Status,
		UpdatedAt:   s.UpdatedAt,
		StartedAt:   s.StartedAt,
		CompletedAt: s.CompletedAt,
	}
}

type stepStormDB struct {
	*storm.DB
}

func newStepStormDB(db *storm.DB) *stepStormDB {
	db.Init(&stormStep{})
	return &stepStormDB{db}
}

func (db *stepStormDB) save(s *Step) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	if err := tx.Save(s.toStormStep()); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func GetStepByID(db *storm.DB, id string) (*Step, error) {
	var sS stormStep
	if err := db.One("ID", id, &sS); err != nil {
		logrus.Error(err)
		return nil, err
	}
	return sS.toStep(), nil
}
