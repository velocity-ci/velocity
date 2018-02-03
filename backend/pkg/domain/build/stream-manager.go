package build

import (
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

type StreamManager struct {
	db *streamDB
}

func NewStreamManager(
	db *gorm.DB,
) *StreamManager {
	db.AutoMigrate(&GormBuild{}, &GormStream{}, &GormStream{})
	m := &StreamManager{
		db: newStreamDB(db),
	}
	return m
}

func (m *StreamManager) new(
	s *Step,
	name string,
) *Stream {
	return &Stream{
		UUID: uuid.NewV3(uuid.NewV1(), s.UUID).String(),
		// Step: s,
		Name: name,
	}
}

func (m *StreamManager) save(s *Stream) error {
	return m.db.save(s)
}
