package domain

import (
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/velocity"
	"gopkg.in/go-playground/validator.v9"
)

type Project struct {
	UUID   string                 `json:"id"`
	Name   string                 `json:"name" validate:"required"`
	Config velocity.GitRepository `json:"repoConfig"`
}

func NewProject(name string, config velocity.GitRepository) (*Project, validator.ValidationErrors) {
	p := &Project{
		Name:   name,
		Config: config,
	}

	err := validate.Struct(p)
	if err != nil {
		return nil, err.(validator.ValidationErrors)
	}

	p.UUID = uuid.NewV1().String()

	return p, nil
}
