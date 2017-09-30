package project

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/middlewares"
)

type requestProject struct {
	ID         string `json:"-"`
	Name       string `json:"name" validate:"required,min=3,max=128,projectUnique"`
	Repository string `json:"repository" validate:"required,min=8,max=128"`
	PrivateKey string `json:"key"`
}

func idFromName(name string) string {
	return strings.Replace(strings.ToLower(name), " ", "-", -1)
}

func FromRequest(b io.ReadCloser) (*domain.Project, error) {
	reqProject := requestProject{}

	err := json.NewDecoder(b).Decode(&reqProject)
	if err != nil {
		return nil, err
	}

	reqProject.ID = idFromName(reqProject.Name)
	reqProject.Name = strings.TrimSpace(reqProject.Name)
	reqProject.Repository = strings.TrimSpace(reqProject.Repository)
	reqProject.PrivateKey = strings.TrimSpace(reqProject.PrivateKey)

	validate, trans := middlewares.GetValidator()

	validate.RegisterValidation("projectUnique", ValidateProjectUnique)
	validate.RegisterTranslation("projectUnique", trans, registerFuncUnique, translationFuncUnique)

	validate.RegisterStructValidation(ValidateProjectRepository, requestProject{})
	validate.RegisterTranslation("repository", trans, registerFuncRepository, translationFuncRepository)
	validate.RegisterTranslation("key", trans, registerFuncKey, translationFuncKey)

	err = validate.Struct(reqProject)

	if err != nil {
		return nil, err
	}

	return &domain.Project{
		ID:            reqProject.ID,
		Name:          reqProject.Name,
		Repository:    reqProject.Repository,
		PrivateKey:    reqProject.PrivateKey,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Synchronising: false,
	}, nil
}
