package domain

import (
	"time"
)

type ResponseProject struct {
	ID string `json:"id"`

	Name       string `json:"name"`
	Repository string `json:"repository"`
	PrivateKey string `json:"key"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Synchronising bool `json:"synchronising"`
}

type Project struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	PrivateKey string `json:"key"`

	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Synchronising bool `json:"synchronising"`
}

func (p *Project) ToResponseProject() *ResponseProject {
	return &ResponseProject{
		ID:            p.ID,
		Name:          p.Name,
		Repository:    p.Repository,
		PrivateKey:    p.PrivateKey,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		Synchronising: p.Synchronising,
	}
}

type Build struct {
	ProjectID  string `json:"projectID"`
	CommitHash string `json:"commitHash"`
}

type Commit struct {
	Branch  string    `json:"branch"`
	Hash    string    `json:"hash"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
	Message string    `json:"message"`
}
