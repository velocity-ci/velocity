package rest

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type projectRequest struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

type projectResponse struct {
}

type projectList struct {
	Total int                `json:"total"`
	Data  []*projectResponse `json:"data"`
}

func newProjectResponse(p *project.Project) *projectResponse {
	return &projectResponse{}
}

type projectHandler struct {
	projectManager *project.Manager
}

func newProjectHandler(projectManager *project.Manager) *projectHandler {
	return &projectHandler{
		projectManager: projectManager,
	}
}

func (h *projectHandler) create(c echo.Context) error {
	rP := new(projectRequest)
	if err := c.Bind(rP); err != nil {
		c.JSON(http.StatusBadRequest, "invalid payload")
		return nil
	}
	p, err := h.projectManager.New(rP.Name, velocity.GitRepository{
		Address:    rP.Address,
		PrivateKey: rP.PrivateKey,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.ErrorMap)
		return nil
	}

	if err := h.projectManager.Save(p); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return nil
	}

	c.JSON(http.StatusCreated, newProjectResponse(p))
	return nil
}
