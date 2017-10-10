package domain

import (
	"fmt"
	"strings"
	"time"
)

type ResponseProject struct {
	ID string `json:"id"`

	Name       string `json:"name"`
	Repository string `json:"repository"`

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

func (c *Commit) OrderedID() string {
	return strings.Join([]string{fmt.Sprintf("%v", c.Date.Unix()), c.Hash[:7]}, "-")
}
