package build

import (
	"encoding/json"
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/golang/glog"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type StormStep struct {
	ID      string `storm:"id"`
	BuildID string `storm:"index"`
	Number  int
	VStep   []byte

	Status      string
	UpdatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
}

func (s *StormStep) toStep(db *storm.DB) *Step {
	var gStep map[string]interface{}
	err := json.Unmarshal(s.VStep, &gStep)
	if err != nil {
		glog.Error(err)
	}

	vStep, err := velocity.DetermineStepFromInterface(gStep)
	if err != nil {
		glog.Error(err)
	} else {
		json.Unmarshal(s.VStep, vStep)
	}
	b, err := GetBuildByID(db, s.BuildID)
	if err != nil {
		glog.Error(err)
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

func (s *Step) toStormStep() *StormStep {
	stepJSON, err := json.Marshal(s.VStep)
	if err != nil {
		glog.Error(err)
	}
	return &StormStep{
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
	db.Init(&StormStep{})
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
	var sS StormStep
	if err := db.One("ID", id, &sS); err != nil {
		glog.Error(err)
		return nil, err
	}
	return sS.toStep(db), nil
}

func getStepsByBuildID(db *storm.DB, buildID string) (r []*Step) {
	query := db.Select(q.Eq("BuildID", buildID)).OrderBy("Number")
	var StormSteps []*StormStep
	query.Find(&StormSteps)

	for _, s := range StormSteps {
		r = append(r, s.toStep(db))
	}

	return r
}
