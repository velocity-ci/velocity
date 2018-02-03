package build

import "github.com/jinzhu/gorm"

type GormStream struct {
	UUID string `gorm:"primary_key"`
	// Step   *GormStep `gorm:"ForeignKey:StepID"`
	// StepID string
	Name string
}

func (GormStream) TableName() string {
	return "build_step_streams"
}

func (g *GormStream) ToStream() *Stream {
	return &Stream{
		UUID: g.UUID,
		// Step: g.Step.ToStep(),
		Name: g.Name,
	}
}

func (s *Stream) ToGormStream() *GormStream {
	return &GormStream{
		UUID: s.UUID,
		// Step: s.Step.ToGormStep(),
		Name: s.Name,
	}
}

type streamDB struct {
	db *gorm.DB
}

func newStreamDB(gorm *gorm.DB) *streamDB {
	return &streamDB{
		db: gorm,
	}
}

func (db *streamDB) save(s *Stream) error {
	tx := db.db.Begin()

	g := s.ToGormStream()

	tx.
		Where(GormStep{UUID: s.UUID}).
		Assign(&g).
		FirstOrCreate(&g)

	return tx.Commit().Error
}
