package build

import (
	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

type StormStream struct {
	ID     string `storm:"id"`
	StepID string `storm:"index"`
	Name   string `json:"name"`
}

func (s *StormStream) toStream(db *storm.DB) *Stream {
	step, err := GetStepByID(db, s.StepID)
	if err != nil {
		logrus.Error(err)
	}
	return &Stream{
		ID:   s.ID,
		Step: step,
		Name: s.Name,
	}
}

func (s *Stream) toStormStream() *StormStream {
	return &StormStream{
		ID:     s.ID,
		StepID: s.Step.ID,
		Name:   s.Name,
	}
}

type streamStormDB struct {
	*storm.DB
}

func newStreamStormDB(db *storm.DB) *streamStormDB {
	db.Init(&StormStep{})
	return &streamStormDB{db}
}

func (db *streamStormDB) save(s *Stream) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	if err := tx.Save(s.toStormStream()); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func getStreamsByStepID(db *storm.DB, stepID string) (r []*Stream) {
	query := db.Select(q.Eq("StepID", stepID))
	var StormStreams []*StormStream
	query.Find(&StormStreams)

	for _, s := range StormStreams {
		r = append(r, s.toStream(db))
	}

	return r
}