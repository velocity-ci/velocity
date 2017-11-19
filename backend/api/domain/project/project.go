package project

import (
	"encoding/json"
	"time"

	"github.com/gosimple/slug"
)

// Repository - Implementing repositories will guarantee consistency between persistence objects and virtual objects.
type Repository interface {
	Save(p *Project) *Project
	Delete(p *Project)
	GetByID(ID string) (*Project, error)
	GetAll(q Query) ([]*Project, uint64)
}
type Project struct {
	ID         string
	Name       string
	Repository GitRepository

	CreatedAt time.Time
	UpdatedAt time.Time

	Synchronising bool
}

type GitRepository struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

type Query struct {
	Amount        uint64
	Page          uint64
	Synchronising bool
}

type ManyResponse struct {
	Total  uint64             `json:"total"`
	Result []*ResponseProject `json:"result"`
}

func (p *Project) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func NewProject(name string, repository GitRepository) *Project {
	return &Project{
		ID:            slug.Make(name),
		Name:          name,
		Repository:    repository,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Synchronising: false,
	}
}

type RequestProject struct {
	ID         string `json:"-"`
	Name       string `json:"name" validate:"required,min=3,max=128,projectUnique"`
	Repository string `json:"repository" validate:"required,min=8,max=128"`
	PrivateKey string `json:"key"`
}

type ResponseProject struct {
	ID string `json:"id"`

	Name       string `json:"name"`
	Repository string `json:"repository"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Synchronising bool `json:"synchronising"`
}

func NewResponseProject(p *Project) *ResponseProject {
	return &ResponseProject{
		ID:            p.ID,
		Name:          p.Name,
		Repository:    p.Repository.Address,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		Synchronising: p.Synchronising,
	}
}
