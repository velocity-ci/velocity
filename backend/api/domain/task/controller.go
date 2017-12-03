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

	// GET /v1/projects/{projectID}/commits/{commitHash}/tasks
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/tasks", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitTasksHandler)),
	)).Methods("GET")

	// GET /v1/projects/{projectID}/commits/{commitHash}/tasks/{taskName}
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/tasks/{taskName}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitTaskHandler)),
	)).Methods("GET")

	c.logger.Println("Set up Task controller.")
}

func (c Controller) getProjectCommitTasksHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]

	_, err := c.projectManager.GetByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	commit, err := c.commitManager.GetCommitByCommitID(reqCommitID)
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

func (c Controller) getProjectCommitTaskHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]
	reqTaskID := reqVars["taskID"]

	_, err := c.projectManager.GetByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	_, err = c.commitManager.GetCommitByCommitID(reqCommitID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	task, err := c.manager.GetByTaskID(reqTaskID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, err)
		return
	}

	c.render.JSON(w, http.StatusOK, NewResponseTask(task))
}