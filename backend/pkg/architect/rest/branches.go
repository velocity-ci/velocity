package rest

import (
	"net/http"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain/project"

	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
)

type branchResponse struct {
	UUID        string    `json:"id"`
	Name        string    `json:"name"`
	LastUpdated time.Time `json:"lastUpdated"`
	Active      bool      `json:"active"`
}

type branchList struct {
	Total int               `json:"total"`
	Data  []*branchResponse `json:"data"`
}

func newBranchResponse(b *githistory.Branch) *branchResponse {
	return &branchResponse{
		UUID:        b.UUID,
		Name:        b.Name,
		LastUpdated: b.LastUpdated,
		Active:      b.Active,
	}
}

type branchHandler struct {
	projectManager *project.Manager
	branchManager  *githistory.BranchManager
}

func newBranchHandler(
	projectManager *project.Manager,
	branchManager *githistory.BranchManager,
) *branchHandler {
	return &branchHandler{
		projectManager: projectManager,
		branchManager:  branchManager,
	}
}

func (h *branchHandler) getAll(c echo.Context) error {

	p := getProjectBySlug(c, h.projectManager)
	if p == nil {
		return nil
	}

	pQ := new(domain.PagingQuery)
	if err := c.Bind(pQ); err != nil {
		c.JSON(http.StatusBadRequest, "invalid parameters")
		return nil
	}

	cs, total := h.branchManager.GetAllForProject(p, pQ)
	rBranches := []*branchResponse{}
	for _, c := range cs {
		rBranches = append(rBranches, newBranchResponse(c))
	}

	c.JSON(http.StatusOK, branchList{
		Total: total,
		Data:  rBranches,
	})

	return nil
}
