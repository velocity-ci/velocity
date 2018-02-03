package rest

import (
	"net/http"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain/project"

	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
)

type commitResponse struct {
	UUID      string    `json:"id"`
	Hash      string    `json:"hash"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
	Message   string    `json:"message"`
	Branches  []string  `json:"branches"`
}

type commitList struct {
	Total int               `json:"total"`
	Data  []*commitResponse `json:"data"`
}

func newCommitResponse(c *githistory.Commit) *commitResponse {
	branches := []string{}
	for _, b := range c.Branches {
		branches = append(branches, b.Name)
	}
	return &commitResponse{
		UUID:      c.UUID,
		Hash:      c.Hash,
		Author:    c.Author,
		CreatedAt: c.CreatedAt,
		Message:   c.Message,
		Branches:  branches,
	}
}

type commitHandler struct {
	projectManager *project.Manager
	commitManager  *githistory.CommitManager
}

func newCommitHandler(
	projectManager *project.Manager,
	commitManager *githistory.CommitManager,
) *commitHandler {
	return &commitHandler{
		projectManager: projectManager,
		commitManager:  commitManager,
	}
}

func (h *commitHandler) getAllForProject(c echo.Context) error {

	p := getProjectBySlug(c, h.projectManager)
	if p == nil {
		return nil
	}

	pQ := new(domain.PagingQuery)
	if err := c.Bind(pQ); err != nil {
		c.JSON(http.StatusBadRequest, "invalid parameters")
		return nil
	}

	cs, total := h.commitManager.GetAllForProject(p, pQ)
	rCommits := []*commitResponse{}
	for _, c := range cs {
		rCommits = append(rCommits, newCommitResponse(c))
	}

	c.JSON(http.StatusOK, commitList{
		Total: total,
		Data:  rCommits,
	})

	return nil
}

func (h *commitHandler) getByProjectAndHash(c echo.Context) error {

	if commit := getCommitByProjectAndHash(c, h.projectManager, h.commitManager); commit != nil {
		c.JSON(http.StatusOK, newCommitResponse(commit))
	}

	return nil
}

func getCommitByProjectAndHash(c echo.Context, pM *project.Manager, cM *githistory.CommitManager) *githistory.Commit {
	p := getProjectBySlug(c, pM)
	if p == nil {
		return nil
	}
	hash := c.Param("hash")
	commit, err := cM.GetByProjectAndHash(p, hash)
	if err != nil {
		c.JSON(http.StatusNotFound, "not found")
		return nil
	}

	return commit
}
