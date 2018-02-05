package build

import (
	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	uuid "github.com/satori/go.uuid"
)

type StreamManager struct {
	db *streamStormDB
}

func NewStreamManager(
	db *storm.DB,
) *StreamManager {
	m := &StreamManager{
		db: newStreamStormDB(db),
	}
	return m
}

func (m *StreamManager) new(
	s *Step,
	name string,
) *Stream {
	return &Stream{
		ID: uuid.NewV3(uuid.NewV1(), s.ID).String(),
		// Step: s,
		Name: name,
	}
}

func (m *StreamManager) save(s *Stream) error {
	return m.db.save(s)
}

func (m *StreamManager) GetByID(id string) (*Stream, error) {
	return GetStreamByID(m.db.DB, id)
}

func GetStreamByID(db *storm.DB, id string) (*Stream, error) {
	var sS stormStream
	if err := db.One("ID", id, &sS); err != nil {
		logrus.Error(err)
		return nil, err
	}
	return sS.toStream(), nil
}
