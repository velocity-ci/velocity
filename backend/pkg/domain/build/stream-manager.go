package build

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
)

type StreamManager struct {
	db          *streamStormDB
	fileManager *StreamFileManager
}

func NewStreamManager(
	db *storm.DB,
	fileManager *StreamFileManager,
) *StreamManager {
	m := &StreamManager{
		db:          newStreamStormDB(db),
		fileManager: fileManager,
	}
	return m
}

func (m *StreamManager) create(
	s *Step,
	name string,
) *Stream {
	stream := &Stream{
		ID: uuid.NewV3(uuid.NewV1(), s.ID).String(),
		// Step: s,
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
	return sL
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
	return sS.toStream(), nil
}
