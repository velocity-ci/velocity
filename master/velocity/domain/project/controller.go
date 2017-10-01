package project

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/middlewares"
)

// Controller - Handles Projects.
type Controller struct {
	logger  *log.Logger
	render  *render.Render
	manager *BoltManager
}

// NewController - Returns a new Controller for Projects.
func NewController(
	controllerLogger *log.Logger,
	renderer *render.Render,
	projectBoltManager *BoltManager,
) *Controller {
	return &Controller{
		logger:  controllerLogger,
		render:  renderer,
		manager: projectBoltManager,
	}
}

// Setup - Sets up the routes for Projects.
func (c Controller) Setup(router *mux.Router) {

	router.Handle("/v1/projects", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectsHandler)),
	)).Methods("GET")

	router.Handle("/v1/projects/{id}", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectHandler)),
	)).Methods("GET")

	router.Handle("/v1/projects/{id}", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.deleteProjectHandler)),
	)).Methods("DELETE")

	router.Handle("/v1/projects", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.postProjectsHandler)),
	)).Methods("POST")

	// POST /v1/projects/{id}/sync
	router.Handle("/v1/projects/{id}/sync", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.syncProjectHandler)),
	)).Methods("POST")

	// GET /v1/projects/{id}/commits
	router.Handle("/v1/projects/{id}/commits", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitsHandler)),
	)).Methods("GET")

	// GET /v1/projects/{projectID}/commits/{commitHash}
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitHandler)),
	)).Methods("GET")

	// GET /v1/projects/{projectID}/commits/{commitHash}/tasks
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/tasks", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitTasksHandler)),
	)).Methods("GET")

	// GET /v1/projects/{projectID}/commits/{commitHash}/tasks/{taskName}
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/tasks/{taskName}", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitTaskHandler)),
	)).Methods("GET")

	c.logger.Println("Set up Project controller.")
}

func (c Controller) getProjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	p, err := c.manager.FindByID(vars["id"])

	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	c.render.JSON(w, http.StatusOK, p.ToResponseProject())
}

func (c Controller) getProjectsHandler(w http.ResponseWriter, r *http.Request) {

	projects := c.manager.FindAll()

	responseProjects := []*domain.ResponseProject{}
	for _, p := range projects {
		responseProjects = append(responseProjects, p.ToResponseProject())
	}

	c.render.JSON(w, http.StatusOK, responseProjects)
}

func (c Controller) deleteProjectHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	err := c.manager.DeleteById(vars["id"])

	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	c.render.JSON(w, http.StatusNoContent, nil)
}

func (c Controller) postProjectsHandler(w http.ResponseWriter, r *http.Request) {
	// username := middlewares.UsernameFromContext(r.Context())

	boltProject, err := FromRequest(r.Body)
	if err != nil {
		middlewares.HandleRequestError(err, w, c.render)
		return
	}

	c.manager.Save(boltProject)

	c.render.JSON(w, http.StatusCreated, boltProject.ToResponseProject())
}

func (c Controller) syncProjectHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["id"]

	p, err := c.manager.FindByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	if p.Synchronising {
		c.render.JSON(w, http.StatusBadRequest, nil)
		return
	}

	p.Synchronising = true
	c.manager.Save(p)
	go sync(p, c.manager)
	c.render.JSON(w, http.StatusCreated, p)
}

func (c Controller) getProjectCommitsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["id"]

	p, err := c.manager.FindByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	commits := c.manager.FindAllCommitsForProject(p)

	c.render.JSON(w, http.StatusOK, commits)
}

func (c Controller) getProjectCommitHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]

	project, err := c.manager.FindByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	commit, err := c.manager.GetCommitInProject(reqCommitID, project)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	c.render.JSON(w, http.StatusOK, commit)
}

func (c Controller) getProjectCommitTasksHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]

	project, err := c.manager.FindByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	commit, err := c.manager.GetCommitInProject(reqCommitID, project)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	tasks := c.manager.GetTasksForCommitInProject(commit, project)

	c.render.JSON(w, http.StatusOK, tasks)
}

func (c Controller) getProjectCommitTaskHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]
	reqTaskName := reqVars["taskName"]

	project, err := c.manager.FindByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	commit, err := c.manager.GetCommitInProject(reqCommitID, project)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	task, err := c.manager.GetTaskForCommitInProject(commit, project, reqTaskName)

	if err != nil {
		c.render.JSON(w, http.StatusNotFound, err)
		return
	}

	c.render.JSON(w, http.StatusOK, task)
}
