package project

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/velocity-ci/velocity/backend/velocity"
)

func NewResolver(projectValidator *Validator) *Resolver {
	return &Resolver{
		projectValidator: projectValidator,
	}
}

type Resolver struct {
	projectValidator *Validator
}

func (r *Resolver) FromRequest(b io.ReadCloser) (Project, error) {
	reqProject := RequestProject{}

	err := json.NewDecoder(b).Decode(&reqProject)
	if err != nil {
		return Project{}, err
	}

	reqProject.Name = strings.TrimSpace(reqProject.Name)
	reqProject.Repository = strings.TrimSpace(reqProject.Repository)
	reqProject.PrivateKey = strings.TrimSpace(reqProject.PrivateKey)

	err = r.projectValidator.Validate(&reqProject)

	if err != nil {
		return Project{}, err
	}

	p := NewProject(
		reqProject.Name,
		velocity.GitRepository{
			Address:    reqProject.Repository,
			PrivateKey: reqProject.PrivateKey,
		},
	)

	return p, nil
}
