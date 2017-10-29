package project

import (
	"encoding/json"
	"time"

	"github.com/velocity-ci/velocity/backend/velocity"
)

type Project struct {
	velocity.Project

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Synchronising bool `json:"synchronising"`
	TotalCommits  uint `json:"totalCommits"`
}

func (p *Project) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func NewProject(name string, repository velocity.GitRepository) Project {
	return Project{
		Project:       velocity.NewProject(name, repository),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Synchronising: false,
		TotalCommits:  0,
	}
}

func (p *Project) ToTaskProject() *velocity.Project {
	return &velocity.Project{
		ID:   p.ID,
		Name: p.Name,
		Repository: velocity.GitRepository{
			Address:    p.Repository.Address,
			PrivateKey: p.Repository.PrivateKey,
		},
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
