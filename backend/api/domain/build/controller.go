package build

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
	"github.com/velocity-ci/velocity/backend/api/middleware"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/backend/api/auth"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
)

// Controller - Handles Builds.
type Controller struct {
	logger         *log.Logger
	render         *render.Render
	manager        Repository
	projectManager project.Repository
	commitManager  commit.Repository
	taskManager    task.Repository
	resolver       *Resolver
}

// NewController - Returns a new Controller for Builds.
func NewController(
	buildResolver *Resolver,
	buildManager Repository,
	projectManager project.Repository,
	commitManager commit.Repository,
	taskManager task.Repository,
) *Controller {
	return &Controller{
		logger:         log.New(os.Stdout, "[controller:build]", log.Lshortfile),
		render:         render.New(),
		manager:        buildManager,
		projectManager: projectManager,
		commitManager:  commitManager,
		taskManager:    taskManager,
		resolver:       buildResolver,
	}
}

// Setup - Sets up the routes for Builds.
func (c Controller) Setup(router *mux.Router) {
	// POST /v1/projects/{id}/commits/{commitID}/builds
	router.Handle("/v1/projects/{projectID}/commits/{commitID}/tasks/{taskID}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.postProjectCommitTaskBuildsHandler)),
	)).Methods("POST")

	// GET /v1/projects/{projectID}/commits/{commitID}/builds
	router.Handle("/v1/projects/{projectID}/commits/{commitID}/tasks/{taskID}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitTaskBuildsHandler)),
	)).Methods("GET")

	// GET /v1/projects/{projectID}/commits/{commitID}/builds/{buildNumber}
	router.Handle("/v1/projects/{projectID}/commits/{commitID}/tasks/{taskID}/builds/{buildNumber}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitTaskBuildHandler)),
	)).Methods("GET")
	// GET /v1/projects/{projectID}/commits/{commitID}/builds/{buildNumber}/steps
	router.Handle("/v1/projects/{projectID}/commits/{commitID}/tasks/{taskID}/builds/{buildNumber}/steps", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitTaskBuildStepsHandler)),
	)).Methods("GET")
	// GET /v1/projects/{projectID}/commits/{commitID}/builds/{buildNumber}/steps/{stepNumber}
	router.Handle("/v1/projects/{projectID}/commits/{commitID}/tasks/{taskID}/builds/{buildNumber}/steps/{stepID}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitTaskBuildStepHandler)),
	)).Methods("GET")
	// GET /v1/projects/{projectID}/commits/{commitID}/builds/{buildNumber}/steps/{stepNumber}/streams
	router.Handle("/v1/projects/{projectID}/commits/{commitID}/tasks/{taskID}/builds/{buildNumber}/steps/{stepID}/streams", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitTaskBuildStepStreamsHandler)),
	)).Methods("GET")
	// GET /v1/projects/{projectID}/commits/{commitID}/builds/{buildNumber}/steps/{stepNumber}/streams/{containerName}
	router.Handle("/v1/projects/{projectID}/commits/{commitID}/tasks/{taskID}/builds/{buildNumber}/steps/{stepID}/streams/{streamName}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitTaskBuildStepStreamHandler)),
	)).Methods("GET")

	c.logger.Println("Set up Commit controller.")
}

func (c Controller) postProjectCommitTaskBuildsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitID"]
	reqTaskID := reqVars["taskID"]

	_, err := c.projectManager.GetByID(reqProjectID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	_, err = c.commitManager.GetCommitByCommitID(reqCommitID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID)) {
		return
	}

	task, err := c.taskManager.GetByTaskID(reqTaskID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find task %s", reqTaskID)) {
		return
	}

	build, err := c.resolver.BuildFromRequest(r.Body, task)
	if err != nil {
		middleware.HandleRequestError(err, w, c.render)
		return
	}

	c.manager.SaveBuild(build)

	c.render.JSON(w, http.StatusCreated, NewResponseBuild(build))

	// queuedBuild := NewQueuedBuild(build, project.ID, commit.Hash)
	// c.manager.QueueBuild(queuedBuild)
}

func (c Controller) getProjectCommitTaskBuildsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitID"]
	reqTaskID := reqVars["taskID"]

	_, err := c.projectManager.GetByID(reqProjectID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	_, err = c.commitManager.GetCommitByCommitID(reqCommitID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID)) {
		return
	}

	task, err := c.taskManager.GetByTaskID(reqTaskID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find task %s", reqTaskID)) {
		return
	}

	builds, count := c.manager.GetBuildsByTaskID(task.ID, Query{})

	respBuilds := []ResponseBuild{}
	for _, b := range builds {
		respBuilds = append(respBuilds, NewResponseBuild(b))
	}

	c.render.JSON(w, http.StatusOK, BuildManyResponse{
		Total:  count,
		Result: respBuilds,
	})
}

func (c Controller) getProjectCommitTaskBuildHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitID"]
	reqTaskID := reqVars["taskID"]
	reqBuildID := reqVars["buildID"]

	_, err := c.projectManager.GetByID(reqProjectID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	_, err = c.commitManager.GetCommitByCommitID(reqCommitID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID)) {
		return
	}

	_, err = c.taskManager.GetByTaskID(reqTaskID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find task %s", reqTaskID)) {
		return
	}

	build, err := c.manager.GetBuildByBuildID(reqBuildID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find build %s", reqBuildID)) {
		return
	}

	c.render.JSON(w, http.StatusOK, NewResponseBuild(build))
}

func (c Controller) getProjectCommitTaskBuildStepsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitID"]
	reqTaskID := reqVars["taskID"]
	reqBuildID := reqVars["buildID"]

	_, err := c.projectManager.GetByID(reqProjectID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	_, err = c.commitManager.GetCommitByCommitID(reqCommitID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID)) {
		return
	}

	t, err := c.taskManager.GetByTaskID(reqTaskID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find task %s", reqTaskID)) {
		return
	}

	build, err := c.manager.GetBuildByBuildID(reqBuildID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find build %s", reqBuildID)) {
		return
	}

	buildSteps, count := c.manager.GetBuildStepsByBuildID(build.ID)
	respBuildSteps := []ResponseBuildStep{}
	for _, buildStep := range buildSteps {
		respBuildSteps = append(respBuildSteps, NewResponseBuildStep(buildStep, t.Steps[buildStep.Number]))
	}
	c.render.JSON(w, http.StatusOK, BuildStepManyResponse{
		Total:  count,
		Result: respBuildSteps,
	})
}

func (c Controller) getProjectCommitTaskBuildStepHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitID"]
	reqTaskID := reqVars["taskID"]
	reqBuildID := reqVars["buildID"]
	reqStepID := reqVars["stepID"]

	_, err := c.projectManager.GetByID(reqProjectID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	_, err = c.commitManager.GetCommitByCommitID(reqCommitID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID)) {
		return
	}

	t, err := c.taskManager.GetByTaskID(reqTaskID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find task %s", reqTaskID)) {
		return
	}

	_, err = c.manager.GetBuildByBuildID(reqBuildID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find build %s", reqBuildID)) {
		return
	}

	buildStep, err := c.manager.GetBuildStepByBuildStepID(reqStepID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find step %s", reqStepID)) {
		return
	}

	c.render.JSON(w, http.StatusOK, NewResponseBuildStep(buildStep, t.Steps[buildStep.Number]))
}

func (c Controller) getProjectCommitTaskBuildStepStreamsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitID"]
	reqTaskID := reqVars["taskID"]
	reqBuildID := reqVars["buildID"]
	reqStepID := reqVars["stepID"]

	_, err := c.projectManager.GetByID(reqProjectID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	_, err = c.commitManager.GetCommitByCommitID(reqCommitID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID)) {
		return
	}

	_, err = c.taskManager.GetByTaskID(reqTaskID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find task %s", reqTaskID)) {
		return
	}

	_, err = c.manager.GetBuildByBuildID(reqBuildID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find build %s", reqBuildID)) {
		return
	}

	buildStep, err := c.manager.GetBuildStepByBuildStepID(reqStepID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find step %s", reqStepID)) {
		return
	}

	outputStreams, count := c.manager.GetStreamsByBuildStepID(buildStep.ID)

	respOutputStreams := []string{}
	for _, outputStream := range outputStreams {
		respOutputStreams = append(respOutputStreams, outputStream.Name)
	}

	c.render.JSON(w, http.StatusOK, OutputStreamManyResponse{
		Total:  count,
		Result: respOutputStreams,
	})
}

func (c Controller) getProjectCommitTaskBuildStepStreamHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitID"]
	reqTaskID := reqVars["taskID"]
	reqBuildID := reqVars["buildID"]
	reqStepID := reqVars["stepID"]
	streamID := reqVars["streamID"]

	_, err := c.projectManager.GetByID(reqProjectID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	_, err = c.commitManager.GetCommitByCommitID(reqCommitID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitID)) {
		return
	}

	_, err = c.taskManager.GetByTaskID(reqTaskID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find task %s", reqTaskID)) {
		return
	}

	_, err = c.manager.GetBuildByBuildID(reqBuildID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find build %s", reqBuildID)) {
		return
	}

	_, err = c.manager.GetBuildStepByBuildStepID(reqStepID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find step %s", reqStepID)) {
		return
	}

	stream, err := c.manager.GetStreamByID(streamID)

	streamLines, total := c.manager.GetStreamLinesByStreamID(stream.ID)

	c.render.JSON(w, http.StatusOK, StreamLineManyResponse{
		Total:  total,
		Result: streamLines,
	})
}

func handleResourceError(r *render.Render, w http.ResponseWriter, err error, message string) bool {
	if err != nil {
		log.Printf(message)
		r.JSON(w, http.StatusNotFound, message)
		return true
	}

	return false
}
