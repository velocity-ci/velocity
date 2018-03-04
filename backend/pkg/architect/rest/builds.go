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

func newBuildResponse(b *build.Build, steps []*stepResponse, branchManager *githistory.BranchManager) *buildResponse {
	return &buildResponse{
		ID:          b.ID,
		Task:        newTaskResponse(b.Task, branchManager),
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

func buildsToBuildResponse(bs []*build.Build, stepManager *build.StepManager, streamManager *build.StreamManager, branchManager *githistory.BranchManager) (r []*buildResponse) {
	for _, b := range bs {
		steps := stepManager.GetStepsForBuild(b)
		rSteps := stepsToStepResponse(steps, streamManager)
		r = append(r, newBuildResponse(b, rSteps, branchManager))
	}
	return r
}

type buildHandler struct {
	buildManager   *build.BuildManager
	stepManager    *build.StepManager
	streamManager  *build.StreamManager
	projectManager *project.Manager
	commitManager  *githistory.CommitManager
	branchManager  *githistory.BranchManager
	taskManager    *task.Manager
}

func newBuildHandler(
	buildManager *build.BuildManager,
	stepManager *build.StepManager,
	streamManager *build.StreamManager,
	projectManager *project.Manager,
	commitManager *githistory.CommitManager,
	branchManager *githistory.BranchManager,
	taskManager *task.Manager,
) *buildHandler {
	return &buildHandler{
		buildManager:   buildManager,
		stepManager:    stepManager,
		streamManager:  streamManager,
		projectManager: projectManager,
		commitManager:  commitManager,
		branchManager:  branchManager,
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

	params := map[string]string{}
	for _, p := range rB.Parameters {
		params[p.Name] = p.Value
	}

	b, err := h.buildManager.Create(t, params)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.ErrorMap)
		return nil
	}

	steps := h.stepManager.GetStepsForBuild(b)
	c.JSON(http.StatusCreated, newBuildResponse(b, stepsToStepResponse(steps, h.streamManager), h.branchManager))
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
	rBuilds := buildsToBuildResponse(bs, h.stepManager, h.streamManager, h.branchManager)

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
	rBuilds := buildsToBuildResponse(bs, h.stepManager, h.streamManager, h.branchManager)

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
	rBuilds := buildsToBuildResponse(bs, h.stepManager, h.streamManager, h.branchManager)

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
	steps := h.stepManager.GetStepsForBuild(b)
	c.JSON(http.StatusOK, newBuildResponse(b, stepsToStepResponse(steps, h.streamManager), h.branchManager))
	return nil
}

func getBuildByID(c echo.Context, buildManager *build.BuildManager) *build.Build {
	id := c.Param("id")
	b, err := buildManager.GetBuildByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, "not found")
		return nil
	}

	return b
}
