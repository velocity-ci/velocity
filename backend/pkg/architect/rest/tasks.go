package rest

import (
	"net/http"

	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/velocity"

	"github.com/velocity-ci/velocity/backend/pkg/domain/project"

	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
)

type taskResponse struct {
	UUID string `json:"id"`
	*velocity.Task
}

type taskList struct {
	Total int             `json:"total"`
	Data  []*taskResponse `json:"data"`
}

func newTaskResponse(t *task.Task) *taskResponse {
	return &taskResponse{
		UUID: t.UUID,
		Task: t.Task,
	}
}

type taskHandler struct {
	projectManager *project.Manager
	commitManager  *githistory.CommitManager
	taskManager    *task.Manager
}

func newTaskHandler(
	projectManager *project.Manager,
	commitManager *githistory.CommitManager,
	taskManager *task.Manager,
) *taskHandler {
	return &taskHandler{
		projectManager: projectManager,
		commitManager:  commitManager,
		taskManager:    taskManager,
	}
}

func (h *taskHandler) getAllForCommit(c echo.Context) error {

	commit := getCommitByProjectAndHash(c, h.projectManager, h.commitManager)
	if commit == nil {
		return nil
	}

	pQ := new(domain.PagingQuery)
	if err := c.Bind(pQ); err != nil {
		c.JSON(http.StatusBadRequest, "invalid parameters")
		return nil
	}

	cs, total := h.taskManager.GetAllForCommit(commit, pQ)
	rTasks := []*taskResponse{}
	for _, c := range cs {
		rTasks = append(rTasks, newTaskResponse(c))
	}

	c.JSON(http.StatusOK, taskList{
		Total: total,
		Data:  rTasks,
	})

	return nil
}

func (h *taskHandler) getByProjectCommitAndName(c echo.Context) error {

	if task := getTaskByCommitAndName(c, h.projectManager, h.commitManager, h.taskManager); task != nil {
		c.JSON(http.StatusOK, newTaskResponse(task))
	}

	return nil
}

func getTaskByCommitAndName(
	c echo.Context,
	pM *project.Manager,
	cM *githistory.CommitManager,
	tM *task.Manager,
) *task.Task {
	commit := getCommitByProjectAndHash(c, pM, cM)
	if c == nil {
		return nil
	}

	name := c.Param("name")
	t, err := tM.GetByCommitAndName(commit, name)
	if err != nil {
		c.JSON(http.StatusNotFound, "not found")
		return nil
	}

	return t
}
