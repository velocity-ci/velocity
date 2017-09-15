package domain

import (
	"time"
)

type Project struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	PrivateKey string `json:"key"`

	ID        string    `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Synchronising bool `json:"synchronising"`

	Builds []Build `json:"builds" gorm:"ForeignKey:ProjectID"`

	Commits []Commit `json:"-" gorm:"-"`
}

type Build struct {
	ProjectID  string `json:"projectID" gorm:"primary_key"`
	CommitHash string `json:"commitHash" gorm:"primary_key"`
}

type Commit struct {
	Hash    string    `json:"hash"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
	Message string    `json:"message"`
}
