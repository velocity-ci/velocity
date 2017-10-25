package project

import (
	"time"

	"github.com/gosimple/slug"
)

type Project struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Repository GitRepository `json:"repository"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Synchronising bool `json:"synchronising"`
	TotalCommits  uint `json:"totalCommits"`
}

func NewProject(name string, repository GitRepository) Project {
	return Project{
		ID:            slug.Make(name),
		Name:          name,
		Repository:    repository,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Synchronising: false,
		TotalCommits:  0,
	}
}

type GitRepository struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
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
