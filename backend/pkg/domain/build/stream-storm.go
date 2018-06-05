package build

import (
	"fmt"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/golang/glog"
)

type StormStream struct {
	ID     string `storm:"id"`
	StepID string `storm:"index"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (s *StormStream) toStream(db *storm.DB) *Stream {
	step, err := GetStepByID(db, s.StepID)
	if err != nil {
		glog.Error(err)
	}
	return &Stream{
		ID:     s.ID,
		Step:   step,
		Name:   s.Name,
		Status: s.Status,
	}
}

func (s *Stream) toStormStream() *StormStream {
	return &StormStream{
		ID:     s.ID,
		StepID: s.Step.ID,
		Name:   s.Name,
		Status: s.Status,
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

type StormStreamLine struct {
	ID         string `storm:"id"`
	StreamID   string `storm:"index"`
	LineNumber int    `storm:"index"`
	Timestamp  time.Time
	Output     string
}

func toStormStreamLine(sL *StreamLine) *StormStreamLine {
	return &StormStreamLine{
		ID:         fmt.Sprintf("%s:%d", sL.StreamID, sL.LineNumber),
		StreamID:   sL.StreamID,
		LineNumber: sL.LineNumber,
		Timestamp:  sL.Timestamp,
		Output:     sL.Output,
	}
}

func toStreamLine(sL *StormStreamLine) *StreamLine {
	return &StreamLine{
		StreamID:   sL.StreamID,
		LineNumber: sL.LineNumber,
		Timestamp:  sL.Timestamp,
		Output:     sL.Output,
	}
}

func (db *streamStormDB) saveStreamLine(streamLine *StreamLine) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	if err := tx.Save(toStormStreamLine(streamLine)); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (db *streamStormDB) getLinesByStream(s *Stream, pQ *domain.PagingQuery) (r []*StreamLine, t int) {
	t = 0
	query := db.Select(q.Eq("StreamID", s.ID))
	t, err := query.Count(&StormStreamLine{})
	if err != nil {
		glog.Error(err)
		return r, t
	}

	var stormStreamLines []*StormStreamLine
	query.OrderBy("LineNumber").Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&stormStreamLines)

	for _, ssL := range stormStreamLines {
		r = append(r, toStreamLine(ssL))
	}

	return r, t

}
