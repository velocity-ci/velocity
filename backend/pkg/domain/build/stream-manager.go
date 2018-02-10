package build

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
)

const (
	EventStreamLineCreate = "streamLine:new"
)

type StreamManager struct {
	db          *streamStormDB
	fileManager *StreamFileManager

	brokers []domain.Broker
}

func NewStreamManager(
	db *storm.DB,
	fileManager *StreamFileManager,
) *StreamManager {
	m := &StreamManager{
		db:          newStreamStormDB(db),
		fileManager: fileManager,
		brokers:     []domain.Broker{},
	}
	return m
}

func (m *StreamManager) AddBroker(b domain.Broker) {
	m.brokers = append(m.brokers, b)
}

func (m *StreamManager) create(
	s *Step,
	name string,
) *Stream {
	stream := &Stream{
		ID:   uuid.NewV3(uuid.NewV1(), s.ID).String(),
		Step: s,
		Name: name,
	}

	m.db.save(stream)
	return stream
}

func (m *StreamManager) update(s *Stream) error {
	return m.db.save(s)
}

func (m *StreamManager) GetByID(id string) (*Stream, error) {
	return GetStreamByID(m.db.DB, id)
}

func (m *StreamManager) CreateStreamLine(
	stream *Stream,
	lineNumber int,
	timestamp time.Time,
	output string,
) *StreamLine {
	sL := &StreamLine{
		StreamID:   stream.ID,
		LineNumber: lineNumber,
		Timestamp:  timestamp,
		Output:     output,
	}
	m.fileManager.saveStreamLine(sL)
	for _, b := range m.brokers {
		b.EmitAll(&domain.Emit{
			Topic:   fmt.Sprintf("stream:%s", stream.ID),
			Event:   EventStreamLineCreate,
			Payload: sL,
		})
	}
	return sL
}

func (m *StreamManager) GetStreamsForStep(s *Step) []*Stream {
	return getStreamsByStepID(m.db.DB, s.ID)
}

func (m *StreamManager) GetStreamLines(s *Stream, q *domain.PagingQuery) ([]*StreamLine, int) {
	return m.fileManager.getLinesByStream(s, q)
}

func GetStreamByID(db *storm.DB, id string) (*Stream, error) {
	var sS stormStream
	if err := db.One("ID", id, &sS); err != nil {
		logrus.Error(err)
		return nil, err
	}
	return sS.toStream(db), nil
}
