package commit

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/backend/api/auth"
	"github.com/velocity-ci/velocity/backend/api/middleware"
	"github.com/velocity-ci/velocity/backend/api/project"
)

// Controller - Handles Projects.
type Controller struct {
	logger         *log.Logger
	render         *render.Render
	manager        *Manager
	projectManager *project.Manager
	resolver       *Resolver
}

// NewController - Returns a new Controller for Projects.
func NewController(
	manager *Manager,
	projectManager *project.Manager,
	commitResolver *Resolver,
) *Controller {
	return &Controller{
		logger:         log.New(os.Stdout, "[controller:commit]", log.Lshortfile),
		render:         render.New(),
		manager:        manager,
		projectManager: projectManager,
		resolver:       commitResolver,
	}
}

// Setup - Sets up the routes for Projects.
func (c Controller) Setup(router *mux.Router) {
	// GET /v1/projects/{id}/commits
	router.Handle("/v1/projects/{id}/commits", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitsHandler)),
	)).Methods("GET")

	// GET /v1/projects/{projectID}/commits/{commitHash}
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitHandler)),
	)).Methods("GET")

	// POST /v1/projects/{id}/sync
	router.Handle("/v1/projects/{id}/sync", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.syncProjectHandler)),
	)).Methods("POST")

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

	// GET /v1/projects/{id}/branches
	router.Handle("/v1/projects/{id}/branches", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getBranchesHandler)),
	)).Methods("GET")

	// POST /v1/projects/{id}/commits/{commitHash}/builds
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.postProjectCommitBuildsHandler)),
	)).Methods("POST")

	c.logger.Println("Set up Commit controller.")
}

func (c Controller) getProjectCommitsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["id"]

	p, err := c.projectManager.FindByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	opts := c.resolver.QueryOptsFromRequest(r)

	commits := c.manager.FindAllCommitsForProject(p, opts)
	totalCommitsForQuery := c.manager.GetTotalCommitsForProject(p, opts)

	c.render.JSON(w, http.StatusOK, CommitsResponse{
		Total:  totalCommitsForQuery,
		Result: commits,
	})
}

func (c Controller) getProjectCommitHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]

	project, err := c.projectManager.FindByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	commit, err := c.manager.GetCommitInProject(reqCommitID, project.ID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	c.render.JSON(w, http.StatusOK, commit)
}

func (c Controller) syncProjectHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["id"]

	p, err := c.projectManager.FindByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	if p.Synchronising {
		c.render.JSON(w, http.StatusBadRequest, map[string][]string{
			"project": []string{"Already synchronising."},
		},
		)
		return
	}

	p.Synchronising = true
	c.projectManager.Save(p)

	go sync(p, c.projectManager, c.manager)

	c.render.JSON(w, http.StatusCreated, p)
}

func (c Controller) getProjectCommitTasksHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]

	project, err := c.projectManager.FindByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	commit, err := c.manager.GetCommitInProject(reqCommitID, project.ID)
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

	project, err := c.projectManager.FindByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	commit, err := c.manager.GetCommitInProject(reqCommitID, project.ID)
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

func (c Controller) getBranchesHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["id"]

	p, err := c.projectManager.FindByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	branches := c.manager.FindAllBranchesForProject(p)
	c.render.JSON(w, http.StatusOK, branches)
}

func (c Controller) postProjectCommitBuildsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]

	project, err := c.projectManager.FindByID(reqProjectID)
	if err != nil {
		log.Printf("Could not find project %s", reqProjectID)
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	commit, err := c.manager.GetCommitInProject(reqCommitID, project.ID)
	if err != nil {
		log.Printf("Could not find commit %s", reqCommitID)
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	build, err := c.resolver.BuildFromRequest(r.Body, project, commit)
	if err != nil {
		middleware.HandleRequestError(err, w, c.render)
		return
	}

	c.manager.SaveBuild(build, project.ID, commit.Hash)

	c.render.JSON(w, http.StatusCreated, NewResponseBuild(build, project.ID, commit.Hash))

	queuedBuild := NewQueuedBuild(build, project.ID, commit.Hash)
	c.manager.QueueBuild(queuedBuild)
}
