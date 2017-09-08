package project

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/velocity-ci/velocity/master/velocity/domain"
)

// FromRequest - Validates and Transforms raw request data into a Project struct.
func FromRequest(b io.ReadCloser) (*domain.Project, error) {
	p := &domain.Project{}

	reqProject := &requestProject{}

	err := json.NewDecoder(b).Decode(reqProject)
	if err != nil {
		return nil, err
	}

	if len(reqProject.Name) < 3 {
		return nil, errors.New("Name too short")
	}

	if len(reqProject.Repository) < 8 {
		return nil, errors.New("Invalid repository address")
	}

	if len(reqProject.Repository) < 8 {
		return nil, errors.New("Invalid key")
	}

	p.ID = strings.Replace(strings.ToLower(reqProject.Name), " ", "-", -1)
	p.Name = reqProject.Name
	p.Repository = reqProject.Repository
	p.PrivateKey = reqProject.PrivateKey

	return p, nil
}

type requestProject struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	PrivateKey string `json:"key"`
}
