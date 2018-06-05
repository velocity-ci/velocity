package project

import (
	"time"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/golang/glog"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type StormProject struct {
	ID            string `storm:"id"`
	Slug          string `storm:"index"`
	Name          string
	Config        velocity.GitRepository
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Synchronising bool
}

func (s *StormProject) ToProject() *Project {
	return &Project{
		ID:            s.ID,
		Slug:          s.Slug,
		Name:          s.Name,
		Config:        s.Config,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
		Synchronising: s.Synchronising,
	}
}

func (p *Project) toStormProject() *StormProject {
	return &StormProject{
		ID:            p.ID,
		Slug:          p.Slug,
		Name:          p.Name,
		Config:        p.Config,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		Synchronising: p.Synchronising,
	}
}

type stormDB struct {
	*storm.DB
}

func newStormDB(db *storm.DB) *stormDB {
	db.Init(&Project{})
	return &stormDB{db}
}

func (db *stormDB) save(p *Project) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	if err := tx.Save(p.toStormProject()); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (db *stormDB) delete(p *Project) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	tx.DeleteStruct(p.toStormProject())

	return tx.Commit()
}

func (db *stormDB) getBySlug(slug string) (*Project, error) {
	query := db.Select(q.Eq("Slug", slug))
	var p StormProject
	if err := query.First(&p); err != nil {
		return nil, err
	}

	return p.ToProject(), nil
}

func (db *stormDB) getByName(name string) (*Project, error) {
	query := db.Select(q.Eq("Name", name))
	var p StormProject
	if err := query.First(&p); err != nil {
		return nil, err
	}

	return p.ToProject(), nil
}

func (db *stormDB) getAll(pQ *domain.PagingQuery) (r []*Project, t int) {
	t = 0
	t, err := db.Count(&StormProject{})
	if err != nil {
		glog.Error(err)
		return r, t
	}

	query := db.Select()
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	var StormProjects []*StormProject
	query.Find(&StormProjects)

	for _, p := range StormProjects {
		r = append(r, p.ToProject())
	}

	return r, t
}

func GetByID(db *storm.DB, id string) (*Project, error) {
	var p StormProject
	if err := db.One("ID", id, &p); err != nil {
		return nil, err
	}
	return p.ToProject(), nil
}
