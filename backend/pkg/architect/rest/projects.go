package rest

import (
	"net/http"
	"time"

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
	ID         string    `json:"id"`
	Slug       string    `json:"slug"`
	Name       string    `json:"name"`
	Repository string    `json:"repository"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`

	Synchronising bool `json:"synchronising"`
}

type projectList struct {
	Total int                `json:"total"`
	Data  []*projectResponse `json:"data"`
}

func newProjectResponse(p *project.Project) *projectResponse {
	return &projectResponse{
		ID:            p.ID,
		Slug:          p.Slug,
		Name:          p.Name,
		Repository:    p.Config.Address,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		Synchronising: p.Synchronising,
	}
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

func (h *projectHandler) getAll(c echo.Context) error {
	pQ := getPagingQueryParams(c)
	if pQ == nil {
		return nil
	}

	ps, total := h.projectManager.GetAll(pQ)
	rProjects := []*projectResponse{}
	for _, p := range ps {
		rProjects = append(rProjects, newProjectResponse(p))
	}

	c.JSON(http.StatusOK, projectList{
		Total: total,
		Data:  rProjects,
	})
	return nil
}

func (h *projectHandler) get(c echo.Context) error {

	if p := getProjectBySlug(c, h.projectManager); p != nil {
		c.JSON(http.StatusOK, newProjectResponse(p))
	}

	return nil
}

func getProjectBySlug(c echo.Context, pM *project.Manager) *project.Project {
	slug := c.Param("slug")

	p, err := pM.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, "not found")
		return nil
	}

	return p
}
