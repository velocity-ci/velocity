package build

import (
	"fmt"
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
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds/{buildNumber}/steps/{stepNumber}/streams/{streamName}", negroni.New(
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
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	commit, err := c.commitManager.GetByProjectAndHash(project, reqCommitID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID)) {
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
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	commit, err := c.commitManager.GetByProjectAndHash(project, reqCommitID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID)) {
		return
	}

	builds, count := c.manager.GetBuildsByProjectAndCommit(project, commit)

	respBuilds := []ResponseBuild{}
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
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	commit, err := c.commitManager.GetByProjectAndHash(project, reqCommitID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID)) {
		return
	}

	build, err := c.manager.GetBuildByProjectAndCommitAndID(project, commit, reqBuildID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find build %s", reqBuildID)) {
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
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	commit, err := c.commitManager.GetByProjectAndHash(project, reqCommitID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID)) {
		return
	}

	build, err := c.manager.GetBuildByProjectAndCommitAndID(project, commit, reqBuildID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find build %s", reqBuildID)) {
		return
	}

	buildSteps, count := c.manager.GetBuildStepsForBuild(build)
	respBuildSteps := []ResponseBuildStep{}
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
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	commit, err := c.commitManager.GetByProjectAndHash(project, reqCommitID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID)) {
		return
	}

	build, err := c.manager.GetBuildByProjectAndCommitAndID(project, commit, reqBuildID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find build %s", reqBuildID)) {
		return
	}

	buildStep, err := c.manager.GetBuildStepByBuildAndID(build, reqStepNumber)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find step %s", reqStepNumber)) {
		return
	}

	c.render.JSON(w, http.StatusOK, NewResponseBuildStep(buildStep))
}

func (c Controller) getProjectCommitBuildStepStreamsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]
	reqBuildID := reqVars["buildID"]
	reqStepNumber := reqVars["stepNumber"]

	project, err := c.projectManager.GetByID(reqProjectID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	commit, err := c.commitManager.GetByProjectAndHash(project, reqCommitID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID)) {
		return
	}

	build, err := c.manager.GetBuildByProjectAndCommitAndID(project, commit, reqBuildID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find build %s", reqBuildID)) {
		return
	}

	buildStep, err := c.manager.GetBuildStepByBuildAndID(build, reqStepNumber)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find step %s", reqStepNumber)) {
		return
	}

	outputStreams, count := c.manager.GetOutputStreamsForBuildStep(buildStep)

	respOutputStreams := []string{}
	for _, outputStream := range outputStreams {
		respOutputStreams = append(respOutputStreams, outputStream.Name)
	}

	c.render.JSON(w, http.StatusOK, OutputStreamManyResponse{
		Total:  count,
		Result: respOutputStreams,
	})
}

func (c Controller) getProjectCommitBuildStepStreamHandler(w http.ResponseWriter, r *http.Request) {
	// project, err := c.projectManager.GetByID(reqProjectID)
	// handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID))

	// commit, err := c.commitManager.GetByProjectAndHash(project, reqCommitID)
	// handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID))

	// build, err := c.manager.GetBuildByProjectAndCommitAndID(project, commit, reqBuildID)
	// handleResourceError(c.render, w, err, fmt.Sprintf("could not find build %s", reqBuildID))

	// buildStep, err := c.manager.GetBuildStepByBuildAndID(build, reqStepNumber)
	// handleResourceError(c.render, w, err, fmt.Sprintf("could not find step %s", reqStepNumber))

	// outputStream, err := c.manager.Get
}

func handleResourceError(r *render.Render, w http.ResponseWriter, err error, message string) bool {
	if err != nil {
		log.Printf(message)
		r.JSON(w, http.StatusNotFound, message)
		return true
	}

	return false
}
