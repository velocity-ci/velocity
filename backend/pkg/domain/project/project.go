package project

import (
	"time"

	"github.com/velocity-ci/velocity/backend/velocity"
)

type Project struct {
	UUID          string                 `json:"id"`
	Slug          string                 `json:"slug"`
	Name          string                 `json:"name" validate:"required,projectUnique"`
	Config        velocity.GitRepository `json:"repoConfig"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
	Synchronising bool                   `json:"synchronising"`
}
