package build

import (
	"log"
	"net/http"
	"os"

	"github.com/velocity-ci/velocity/backend/api/domain/commit"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/backend/api/auth"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/api/middleware"
)

// Controller - Handles Builds.
type Controller struct {
	logger         *log.Logger
	render         *render.Render
	manager        Repository
	projectManager project.Repository
	commitManager  commit.Repository
	// resolver       *Resolver
}

// NewController - Returns a new Controller for Builds.
func NewController(
	manager *Manager,
) *Controller {
	return &Controller{
		logger:  log.New(os.Stdout, "[controller:commit]", log.Lshortfile),
		render:  render.New(),
		manager: manager,
	}
}

// Setup - Sets up the routes for Builds.
func (c Controller) Setup(router *mux.Router) {
	// POST /v1/projects/{id}/commits/{commitHash}/builds
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.postProjectCommitBuildsHandler)),
	)).Methods("POST")

	// GET /v1/projects/{id}/commits/{commitHash}/builds
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitBuildsHandler)),
	)).Methods("GET")

	c.logger.Println("Set up Commit controller.")
}

func (c Controller) postProjectCommitBuildsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]

	project, err := c.projectManager.GetByID(reqProjectID)
	if err != nil {
		log.Printf("Could not find project %s", reqProjectID)
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	commit, err := c.commitManager.GetByProjectAndHash(project, reqCommitID)
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

	c.manager.SaveToProjectAndCommit(project, commit, build)

	c.render.JSON(w, http.StatusCreated, NewResponseBuild(build, project.ID, commit.Hash))

	queuedBuild := NewQueuedBuild(build, project.ID, commit.Hash)
	c.manager.QueueBuild(queuedBuild)
}

func (c Controller) getProjectCommitBuildsHandler(w http.ResponseWriter, r *http.Request) {
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

	builds := c.manager.GetBuilds(project.ID, commit.Hash)

	respBuilds := []*ResponseBuild{}
	for _, b := range builds {
		respBuilds = append(respBuilds, NewResponseBuild(b, project.ID, commit.Hash))
	}

	c.render.JSON(w, http.StatusOK, respBuilds)
}
