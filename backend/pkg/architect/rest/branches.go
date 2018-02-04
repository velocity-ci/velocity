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
	commitManager  *githistory.CommitManager
}

func newBranchHandler(
	projectManager *project.Manager,
	branchManager *githistory.BranchManager,
	commitManager *githistory.CommitManager,
) *branchHandler {
	return &branchHandler{
		projectManager: projectManager,
		branchManager:  branchManager,
		commitManager:  commitManager,
	}
}

func (h *branchHandler) getAllForProject(c echo.Context) error {

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

func (h *branchHandler) getByProjectAndName(c echo.Context) error {

	if b := getBranchByProjectAndName(c, h.projectManager, h.branchManager); b != nil {
		c.JSON(http.StatusOK, newBranchResponse(b))
	}

	return nil
}

func getBranchByProjectAndName(c echo.Context, pM *project.Manager, bM *githistory.BranchManager) *githistory.Branch {
	p := getProjectBySlug(c, pM)
	if p == nil {
		return nil
	}
	name := c.Param("name")
	b, err := bM.GetByProjectAndName(p, name)
	if err != nil {
		c.JSON(http.StatusNotFound, "not found")
		return nil
	}

	return b
}

func (h *branchHandler) getCommitsForBranch(c echo.Context) error {

	b := getBranchByProjectAndName(c, h.projectManager, h.branchManager)

	pQ := new(domain.PagingQuery)
	if err := c.Bind(pQ); err != nil {
		c.JSON(http.StatusBadRequest, "invalid parameters")
		return nil
	}

	cs, total := h.commitManager.GetAllForBranch(b, pQ)
	rCommits := []*commitResponse{}
	for _, c := range cs {
		bs, _ := h.branchManager.GetAllForCommit(c, &domain.PagingQuery{Limit: 100, Page: 1})
		rCommits = append(rCommits, newCommitResponse(c, bs))
	}

	c.JSON(http.StatusOK, commitList{
		Total: total,
		Data:  rCommits,
	})

	return nil
}
