package web

import (
	"net/http"

	"github.com/velocity-ci/velocity/backend/velocity"

	"github.com/velocity-ci/velocity/backend/architect/domain"
	"github.com/velocity-ci/velocity/backend/architect/domain/persistence"

	"github.com/labstack/echo"
)

type RequestProject struct {
	Name        string `json:"name"`
	RepoAddress string `json:"repositoryAddress"`
}

type ResponseProject struct {
	UUID        string `json:"id"`
	Name        string `json:"name"`
	RepoAddress string `json:"repositoryAddress"`
}

func NewResponseProject(p *domain.Project) *ResponseProject {
	return &ResponseProject{
		UUID: p.UUID,
		Name: p.Name,
	}
}

func createProject(c echo.Context) error {
	rP := new(RequestProject)
	if err := c.Bind(rP); err != nil {
		c.JSON(http.StatusBadRequest, "invalid payload")
		c.Logger().Warn(err)
		return nil
	}

	p, err := domain.NewProject(rP.Name, velocity.GitRepository{
		Address: rP.RepoAddress,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorMap(err))
	}

	if err := persistence.SaveProject(p); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return nil
	}

	c.JSON(http.StatusCreated, NewResponseProject(p))
	return nil
}
