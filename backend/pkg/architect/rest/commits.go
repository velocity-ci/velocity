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
	ID        string    `json:"id"`
	Hash      string    `json:"hash"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
	Message   string    `json:"message"`
	Signed    string    `json:"signed"`
	Branches  []string  `json:"branches"`
}

type commitList struct {
	Total int               `json:"total"`
	Data  []*commitResponse `json:"data"`
}

func newCommitResponse(c *githistory.Commit, bs []*githistory.Branch) *commitResponse {
	branches := []string{}
	for _, b := range bs {
		branches = append(branches, b.Name)
	}
	return &commitResponse{
		ID:        c.ID,
		Hash:      c.Hash,
		Author:    c.Author,
		CreatedAt: c.CreatedAt,
		Message:   c.Message,
		Signed:    c.Signed,
		Branches:  branches,
	}
}

type commitHandler struct {
	projectManager *project.Manager
	commitManager  *githistory.CommitManager
	branchManager  *githistory.BranchManager
}

func newCommitHandler(
	projectManager *project.Manager,
	commitManager *githistory.CommitManager,
	branchManager *githistory.BranchManager,
) *commitHandler {
	return &commitHandler{
		projectManager: projectManager,
		commitManager:  commitManager,
		branchManager:  branchManager,
	}
}

func getCommitQueryParams(c echo.Context) *githistory.CommitQuery {
	pQ := githistory.NewCommitQuery()
	if err := c.Bind(pQ); err != nil {
		c.JSON(http.StatusBadRequest, "invalid parameters")
		return nil
	}

	if pQ.Branch != "" {
		pQ.Branches = []string{pQ.Branch}
	}

	return pQ
}

func (h *commitHandler) getAllForProject(c echo.Context) error {

	p := getProjectBySlug(c, h.projectManager)
	if p == nil {
		return nil
	}

	pQ := getCommitQueryParams(c)
	if pQ == nil {
		return nil
	}

	cs, total := h.commitManager.GetAllForProject(p, pQ)
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

func (h *commitHandler) getByProjectAndHash(c echo.Context) error {

	if commit := getCommitByProjectAndHash(c, h.projectManager, h.commitManager); commit != nil {
		bs, _ := h.branchManager.GetAllForCommit(commit, &domain.PagingQuery{Limit: 100, Page: 1})
		c.JSON(http.StatusOK, newCommitResponse(commit, bs))
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
