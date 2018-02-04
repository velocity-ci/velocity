package build

import "github.com/asdine/storm"

type stormStream struct {
	ID   string `storm:"id"`
	Name string `json:"name"`
}

func (s *stormStream) toStream() *Stream {
	return &Stream{
		UUID: s.ID,
		Name: s.Name,
	}
}

func (s *Stream) toStormStream() stormStream {
	return stormStream{
		ID:   s.UUID,
		Name: s.Name,
	}
}

type streamStormDB struct {
	*storm.DB
}

func newStreamStormDB(db *storm.DB) *streamStormDB {
	db.Init(&stormStep{})
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
