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
	resolver       *Resolver
}

// NewController - Returns a new Controller for Builds.
func NewController(
	buildResolver *Resolver,
	buildManager *Manager,
	projectManager *project.Manager,
	commitManager *commit.Manager,
) *Controller {
	return &Controller{
		logger:         log.New(os.Stdout, "[controller:commit]", log.Lshortfile),
		render:         render.New(),
		manager:        buildManager,
		projectManager: projectManager,
		commitManager:  commitManager,
		resolver:       buildResolver,
	}
}

// Setup - Sets up the routes for Builds.
func (c Controller) Setup(router *mux.Router) {
	// POST /v1/projects/{id}/commits/{commitHash}/builds
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.postProjectCommitBuildsHandler)),
	)).Methods("POST")

	// GET /v1/projects/{projectID}/commits/{commitHash}/builds
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitBuildsHandler)),
	)).Methods("GET")

	// GET /v1/projects/{projectID}/commits/{commitHash}/builds/{buildNumber}
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds/{buildNumber}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitBuildHandler)),
	)).Methods("GET")
	// GET /v1/projects/{projectID}/commits/{commitHash}/builds/{buildNumber}/steps
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds/{buildNumber}/steps", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitBuildStepsHandler)),
	)).Methods("GET")
	// GET /v1/projects/{projectID}/commits/{commitHash}/builds/{buildNumber}/steps/{stepNumber}
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds/{buildNumber}/steps/{stepNumber}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitBuildStepHandler)),
	)).Methods("GET")
	// GET /v1/projects/{projectID}/commits/{commitHash}/builds/{buildNumber}/steps/{stepNumber}/streams
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds/{buildNumber}/steps/{stepNumber}/streams", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitBuildStepStreamsHandler)),
	)).Methods("GET")
	// GET /v1/projects/{projectID}/commits/{commitHash}/builds/{buildNumber}/steps/{stepNumber}/streams/{containerName}
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds/{buildNumber}/steps/{stepNumber}/streams/{countainerName}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitBuildStepStreamHandler)),
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

	c.manager.SaveBuild(build)

	c.render.JSON(w, http.StatusCreated, NewResponseBuild(build))

	// queuedBuild := NewQueuedBuild(build, project.ID, commit.Hash)
	// c.manager.QueueBuild(queuedBuild)
}

func (c Controller) getProjectCommitBuildsHandler(w http.ResponseWriter, r *http.Request) {
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

	builds, count := c.manager.GetBuildsByProjectAndCommit(project, commit)

	respBuilds := []*ResponseBuild{}
	for _, b := range builds {
		respBuilds = append(respBuilds, NewResponseBuild(b))
	}

	c.render.JSON(w, http.StatusOK, BuildManyResponse{
		Total:  count,
		Result: respBuilds,
	})
}

func (c Controller) getProjectCommitBuildHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]
	reqBuildID := reqVars["buildID"]

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

	build, err := c.manager.GetBuildByProjectAndCommitAndID(project, commit, reqBuildID)
	if err != nil {
		log.Printf("Could not find build %s", reqBuildID)
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	c.render.JSON(w, http.StatusOK, NewResponseBuild(build))
}

func (c Controller) getProjectCommitBuildStepsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]
	reqBuildID := reqVars["buildID"]

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

	build, err := c.manager.GetBuildByProjectAndCommitAndID(project, commit, reqBuildID)
	if err != nil {
		log.Printf("Could not find build %s", reqBuildID)
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	buildSteps, count := c.manager.GetBuildStepsForBuild(build)
	respBuildSteps := []*ResponseBuildStep{}
	for _, buildStep := range buildSteps {
		respBuildSteps = append(respBuildSteps, NewResponseBuildStep(buildStep))
	}
	c.render.JSON(w, http.StatusOK, BuildStepManyResponse{
		Total:  count,
		Result: respBuildSteps,
	})
}

func (c Controller) getProjectCommitBuildStepHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]
	reqBuildID := reqVars["buildID"]
	reqStepNumber := reqVars["stepNumber"]

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

	build, err := c.manager.GetBuildByProjectAndCommitAndID(project, commit, reqBuildID)
	if err != nil {
		log.Printf("Could not find build %s", reqBuildID)
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	buildStep, err := c.manager.GetBuildStepByBuildAndID(build, reqStepNumber)
	if err != nil {
		log.Printf("Could not find step %s", reqStepNumber)
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	c.render.JSON(w, http.StatusOK, NewResponseBuildStep(buildStep))
}

func (c Controller) getProjectCommitBuildStepStreamsHandler(w http.ResponseWriter, r *http.Request) {

}

func (c Controller) getProjectCommitBuildStepStreamHandler(w http.ResponseWriter, r *http.Request) {

}
