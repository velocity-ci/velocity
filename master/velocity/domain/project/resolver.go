package project

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/middlewares"
)

func FromRequest(b io.ReadCloser) (*domain.Project, error) {
	p := &domain.Project{}

	err := json.NewDecoder(b).Decode(p)
	if err != nil {
		return nil, err
	}

	p.ID = strings.Replace(strings.ToLower(p.Name), " ", "-", -1)

	return p, nil
}

func ValidatePOST(p *domain.Project, dbManager *DBManager) (bool, *middlewares.ResponseErrors) {
	hasErrors := false

	errs := projectErrors{}

	if len(p.Name) < 3 || len(p.Name) > 128 {
		errs.Name = "Invalid name"
		hasErrors = true
	}

	if len(p.Repository) < 8 || len(p.Repository) > 128 {
		errs.Repository = "Invalid repository address"
		hasErrors = true
	}

	if len(p.PrivateKey) < 8 {
		errs.PrivateKey = "Invalid key"
		hasErrors = true
	}

	if hasErrors {
		return false, &middlewares.ResponseErrors{
			Errors: &errs,
		}
	}
	_, err := dbManager.FindByID(p.ID)
	if err == nil {
		return false, &middlewares.ResponseErrors{
			Errors: &projectErrors{
				Name: "Name already taken.",
			},
		}
	}

	return true, nil
}

type projectErrors struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	PrivateKey string `json:"key"`
}
