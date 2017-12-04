package task

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/backend/api/auth"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
)

// Controller - Handles Tasks.
type Controller struct {
	logger         *log.Logger
	render         *render.Render
	manager        Repository
	projectManager project.Repository
	commitManager  commit.Repository
}

// NewController - Returns a new Controller for Tasks.
func NewController(
	manager *Manager,
	projectManager *project.Manager,
	commitManager *commit.Manager,
) *Controller {
	return &Controller{
		logger:         log.New(os.Stdout, "[controller:task]", log.Lshortfile),
		render:         render.New(),
		manager:        manager,
		projectManager: projectManager,
		commitManager:  commitManager,
	}
}

// Setup - Sets up the routes for Tasks.
func (c Controller) Setup(router *mux.Router) {

	// GET /v1/commits/{commitUUID}/tasks
	router.Handle("/v1/commits/{commitUUID}/tasks", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getTasksByCommitUUIDHandler)),
	)).Methods("GET")

	// GET /v1/projects/{projectID}/commits/{commitHash}/tasks
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/tasks", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitTasksHandler)),
	)).Methods("GET")

	// GET /v1/tasks/{taskUUID}
	router.Handle("/v1/tasks/{taskUUID}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getTaskByUUIDHandler)),
	)).Methods("GET")

	// GET /v1/projects/{projectID}/commits/{commitHash}/tasks/{taskName}
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/tasks/{taskName}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitTaskHandler)),
	)).Methods("GET")

	c.logger.Println("Set up Task controller.")
}

func (c Controller) getTasksByCommitUUIDHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqCommitUUID := reqVars["commitUUID"]

	commit, err := c.commitManager.GetCommitByCommitID(reqCommitUUID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	tasks, count := c.manager.GetAllByCommitID(commit.ID, Query{})

	responseTasks := []ResponseTask{}

	for _, t := range tasks {
		responseTasks = append(responseTasks, NewResponseTask(t))
	}

	c.render.JSON(w, http.StatusOK, ManyResponse{
		Total:  count,
		Result: responseTasks,
	})
}

func (c Controller) getProjectCommitTasksHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitHash := reqVars["commitHash"]

	p, err := c.projectManager.GetByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	commit, err := c.commitManager.GetCommitByProjectIDAndCommitHash(p.ID, reqCommitHash)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	tasks, count := c.manager.GetAllByCommitID(commit.ID, Query{})

	responseTasks := []ResponseTask{}

	for _, t := range tasks {
		responseTasks = append(responseTasks, NewResponseTask(t))
	}

	c.render.JSON(w, http.StatusOK, ManyResponse{
		Total:  count,
		Result: responseTasks,
	})
}

func (c Controller) getTaskByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqTaskUUID := reqVars["taskUUID"]

	task, err := c.manager.GetByTaskID(reqTaskUUID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, err)
		return
	}

	c.render.JSON(w, http.StatusOK, NewResponseTask(task))
}

func (c Controller) getProjectCommitTaskHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitHash := reqVars["commitHash"]
	reqTaskName := reqVars["taskName"]

	p, err := c.projectManager.GetByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	cm, err := c.commitManager.GetCommitByProjectIDAndCommitHash(p.ID, reqCommitHash)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	task, err := c.manager.GetByCommitIDAndTaskName(cm.ID, reqTaskName)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, err)
		return
	}

	c.render.JSON(w, http.StatusOK, NewResponseTask(task))
}
