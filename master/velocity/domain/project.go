package domain

import (
	"time"
)

type Project struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	PrivateKey string `json:"key"`

	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Synchronising bool `json:"synchronising"`

	Builds []Build `json:"builds"`
}

type Build struct {
	ProjectID  string `json:"projectID"`
	CommitHash string `json:"commitHash"`
}

type Commit struct {
	Hash    string    `json:"hash"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
	Message string    `json:"message"`
}
