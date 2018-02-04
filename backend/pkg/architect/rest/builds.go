package rest

import (
	"net/http"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"

	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type buildRequest struct {
	Parameters []requestParameter `json:"params"`
}

type requestParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type buildResponse struct {
	ID   string        `json:"id"`
	Task *taskResponse `json:"task"`

	Steps []*stepResponse `json:"buildSteps"`

	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt"`
}

func newBuildResponse(b *build.Build) *buildResponse {
	steps := []*stepResponse{}
	for _, s := range b.Steps {
		steps = append(steps, newStepResponse(s))
	}
	return &buildResponse{
		ID:          b.ID,
		Task:        newTaskResponse(b.Task),
		Steps:       steps,
		Status:      b.Status,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
		StartedAt:   b.StartedAt,
		CompletedAt: b.CompletedAt,
	}
}

type buildList struct {
	Total int              `json:"total"`
	Data  []*buildResponse `json:"data"`
}

func buildsToBuildResponse(bs []*build.Build) (r []*buildResponse) {
	for _, b := range bs {
		r = append(r, newBuildResponse(b))
	}
	return r
}

type buildHandler struct {
	buildManager   *build.BuildManager
	projectManager *project.Manager
	commitManager  *githistory.CommitManager
	taskManager    *task.Manager
}

func newBuildHandler(
	buildManager *build.BuildManager,
	projectManager *project.Manager,
	commitManager *githistory.CommitManager,
	taskManager *task.Manager,
) *buildHandler {
	return &buildHandler{
		buildManager:   buildManager,
		projectManager: projectManager,
		commitManager:  commitManager,
		taskManager:    taskManager,
	}
}

func (h *buildHandler) create(c echo.Context) error {
	rB := new(buildRequest)
	if err := c.Bind(rB); err != nil {
		c.JSON(http.StatusBadRequest, "invalid payload")
		return nil
	}

	t := getTaskByCommitAndSlug(c, h.projectManager, h.commitManager, h.taskManager)
	if t == nil {
		return nil
	}

	b, err := h.buildManager.New(t, map[string]string{})
	if err != nil {
		c.JSON(http.StatusBadRequest, err.ErrorMap)
		return nil
	}

	if err := h.buildManager.Save(b); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return nil
	}

	c.JSON(http.StatusCreated, newBuildResponse(b))
	return nil
}

func (h *buildHandler) getAllForProject(c echo.Context) error {

	p := getProjectBySlug(c, h.projectManager)
	if p == nil {
		return nil
	}

	pQ := getPagingQueryParams(c)
	if pQ == nil {
		return nil
	}

	bs, total := h.buildManager.GetAllForProject(p, pQ)
	rBuilds := buildsToBuildResponse(bs)

	c.JSON(http.StatusOK, buildList{
		Total: total,
		Data:  rBuilds,
	})

	return nil
}

func (h *buildHandler) getAllForCommit(c echo.Context) error {

	commit := getCommitByProjectAndHash(c, h.projectManager, h.commitManager)
	if commit == nil {
		return nil
	}

	pQ := getPagingQueryParams(c)
	if pQ == nil {
		return nil
	}

	bs, total := h.buildManager.GetAllForCommit(commit, pQ)
	rBuilds := buildsToBuildResponse(bs)

	c.JSON(http.StatusOK, buildList{
		Total: total,
		Data:  rBuilds,
	})

	return nil
}

func (h *buildHandler) getAllForTask(c echo.Context) error {

	t := getTaskByCommitAndSlug(c, h.projectManager, h.commitManager, h.taskManager)
	if t == nil {
		return nil
	}

	pQ := getPagingQueryParams(c)
	if pQ == nil {
		return nil
	}

	bs, total := h.buildManager.GetAllForTask(t, pQ)
	rBuilds := buildsToBuildResponse(bs)

	c.JSON(http.StatusOK, buildList{
		Total: total,
		Data:  rBuilds,
	})

	return nil
}

func (h *buildHandler) getByID(c echo.Context) error {
	b := getBuildByID(c, h.buildManager)
	if b == nil {
		return nil
	}
	c.JSON(http.StatusOK, newBuildResponse(b))
	return nil
}

func getBuildByID(c echo.Context, buildManager *build.BuildManager) *build.Build {
	id := c.Param("id")
	b, err := buildManager.GetBuildByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, "not found")
		return nil
	}
}
