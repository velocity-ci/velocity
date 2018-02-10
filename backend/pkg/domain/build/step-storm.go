package build

import (
	"encoding/json"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type stormStep struct {
	ID      string `storm:"id"`
	BuildID string `storm:"index"`
	Number  int
	VStep   []byte

	Status      string
	UpdatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
}

func (s *stormStep) toStep(db *storm.DB) *Step {
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
	b, err := GetBuildByID(db, s.BuildID)
	if err != nil {
		logrus.Error(err)
	}
	return &Step{
		ID:          s.ID,
		Build:       b,
		Number:      s.Number,
		Status:      s.Status,
		UpdatedAt:   s.UpdatedAt,
		StartedAt:   s.StartedAt,
		CompletedAt: s.CompletedAt,
		VStep:       &vStep,
	}
}

func (s *Step) toStormStep() *stormStep {
	stepJSON, err := json.Marshal(s.VStep)
	if err != nil {
		logrus.Error(err)
	}
	return &stormStep{
		ID:          s.ID,
		BuildID:     s.Build.ID,
		Number:      s.Number,
		VStep:       stepJSON,
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
	return sS.toStep(db), nil
}

func getStepsByBuildID(db *storm.DB, buildID string) (r []*Step) {
	query := db.Select(q.Eq("BuildID", buildID)).OrderBy("Number")
	var stormSteps []*stormStep
	query.Find(&stormSteps)

	for _, s := range stormSteps {
		r = append(r, s.toStep(db))
	}

	return r
}
