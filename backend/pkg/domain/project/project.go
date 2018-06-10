package project

import (
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type Project struct {
	ID            string                 `json:"id"`
	Slug          string                 `json:"slug"`
	Name          string                 `json:"name" validate:"required,projectUnique"`
	Config        velocity.GitRepository `json:"repoConfig"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
	Synchronising bool                   `json:"synchronising"`

	velocity.RepositoryConfig
}
