package build

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/backend/api/auth"
	"github.com/velocity-ci/velocity/backend/api/middleware"
)

func (c Controller) addBuildRoutes(router *mux.Router) {
	// POST /v1/projects/{id}/commits/{commitHash}/tasks/{taskName}/builds
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/tasks/{taskName}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.postProjectCommitTaskBuildsHandler)),
	)).Methods("POST")

	// POST /v1/tasks/{taskUUID}/builds
	router.Handle("/v1/tasks/{taskUUID}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.postTaskByUUIDBuildsHandler)),
	)).Methods("POST")

	// GET /v1/projects/{projectID}/builds
	router.Handle("/v1/projects/{projectID}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectBuildsHandler)),
	)).Methods("GET")

	// GET /v1/projects/{projectID}/commits/{commitHash}/builds
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitBuildsHandler)),
	)).Methods("GET")

	// GET /v1/projects/{projectID}/commits/{commitHash}/tasks/{taskName}/builds
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}/tasks/{taskName}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitTaskBuildsHandler)),
	)).Methods("GET")

	// GET /v1/commits/{commitUUID}/builds
	router.Handle("/v1/commits/{commitUUID}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getCommitByUUIDBuildsHandler)),
	)).Methods("GET")

	// GET /v1/tasks/{taskUUID}/builds
	router.Handle("/v1/tasks/{taskUUID}/builds", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getTaskByUUIDBuildsHandler)),
	)).Methods("GET")

	// GET /v1/builds/{buildUUID}
	router.Handle("/v1/builds/{buildUUID}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getBuildByUUIDHandler)),
	)).Methods("GET")
}

func (c Controller) postProjectCommitTaskBuildsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitHash := reqVars["commitHash"]
	reqTaskName := reqVars["taskName"]

	p, err := c.projectManager.GetByID(reqProjectID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	cm, err := c.commitManager.GetCommitByProjectIDAndCommitHash(p.ID, reqCommitHash)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitHash)) {
		return
	}

	task, err := c.taskManager.GetByCommitIDAndTaskName(cm.ID, reqTaskName)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find task %s", reqTaskName)) {
		return
	}

	build, err := c.resolver.BuildFromRequest(r.Body, task)
	if err != nil {
		middleware.HandleRequestError(err, w, c.render)
		return
	}

	build = c.manager.CreateBuild(build)

	c.render.JSON(w, http.StatusCreated, NewResponseBuild(build))

	// queuedBuild := NewQueuedBuild(build, project.ID, commit.Hash)
	// c.manager.QueueBuild(queuedBuild)
}

func (c Controller) postTaskByUUIDBuildsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqTaskUUID := reqVars["taskUUID"]

	task, err := c.taskManager.GetByTaskID(reqTaskUUID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find task %s", reqTaskUUID)) {
		return
	}

	build, err := c.resolver.BuildFromRequest(r.Body, task)
	if err != nil {
		middleware.HandleRequestError(err, w, c.render)
		return
	}

	build = c.manager.CreateBuild(build)

	c.render.JSON(w, http.StatusCreated, NewResponseBuild(build))
}

func (c Controller) getProjectBuildsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]

	p, err := c.projectManager.GetByID(reqProjectID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	opts := BuildQueryOptsFromRequest(r)

	builds, count := c.manager.GetBuildsByProjectID(p.ID, opts)

	respBuilds := []ResponseBuild{}
	for _, b := range builds {
		respBuilds = append(respBuilds, NewResponseBuild(b))
	}

	c.render.JSON(w, http.StatusOK, BuildManyResponse{
		Total:  count,
		Result: respBuilds,
	})
}

func (c Controller) getProjectCommitBuildsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitHash := reqVars["commitHash"]

	p, err := c.projectManager.GetByID(reqProjectID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	cm, err := c.commitManager.GetCommitByProjectIDAndCommitHash(p.ID, reqCommitHash)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitHash)) {
		return
	}

	opts := BuildQueryOptsFromRequest(r)
	builds, count := c.manager.GetBuildsByCommitID(cm.ID, opts)

	respBuilds := []ResponseBuild{}
	for _, b := range builds {
		respBuilds = append(respBuilds, NewResponseBuild(b))
	}
	c.render.JSON(w, http.StatusOK, BuildManyResponse{
		Total:  count,
		Result: respBuilds,
	})
}

func (c Controller) getProjectCommitTaskBuildsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitHash := reqVars["commitHash"]
	reqTaskName := reqVars["taskName"]

	p, err := c.projectManager.GetByID(reqProjectID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find project %s", reqProjectID)) {
		return
	}

	cm, err := c.commitManager.GetCommitByProjectIDAndCommitHash(p.ID, reqCommitHash)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitHash)) {
		return
	}

	task, err := c.taskManager.GetByCommitIDAndTaskName(cm.ID, reqTaskName)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find task %s", reqTaskName)) {
		return
	}

	opts := BuildQueryOptsFromRequest(r)

	builds, count := c.manager.GetBuildsByTaskID(task.ID, opts)

	respBuilds := []ResponseBuild{}
	for _, b := range builds {
		respBuilds = append(respBuilds, NewResponseBuild(b))
	}

	c.render.JSON(w, http.StatusOK, BuildManyResponse{
		Total:  count,
		Result: respBuilds,
	})
}

func (c Controller) getCommitByUUIDBuildsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqCommitUUID := reqVars["commitUUID"]

	cm, err := c.commitManager.GetCommitByCommitID(reqCommitUUID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find commit %s", reqCommitUUID)) {
		return
	}

	opts := BuildQueryOptsFromRequest(r)

	builds, count := c.manager.GetBuildsByCommitID(cm.ID, opts)

	respBuilds := []ResponseBuild{}
	for _, b := range builds {
		respBuilds = append(respBuilds, NewResponseBuild(b))
	}

	c.render.JSON(w, http.StatusOK, BuildManyResponse{
		Total:  count,
		Result: respBuilds,
	})
}

func (c Controller) getTaskByUUIDBuildsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqTaskUUID := reqVars["taskUUID"]

	task, err := c.taskManager.GetByTaskID(reqTaskUUID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find task %s", reqTaskUUID)) {
		return
	}

	opts := BuildQueryOptsFromRequest(r)

	builds, count := c.manager.GetBuildsByTaskID(task.ID, opts)

	respBuilds := []ResponseBuild{}
	for _, b := range builds {
		respBuilds = append(respBuilds, NewResponseBuild(b))
	}

	c.render.JSON(w, http.StatusOK, BuildManyResponse{
		Total:  count,
		Result: respBuilds,
	})
}

func (c Controller) getBuildByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqBuildUUID := reqVars["buildUUID"]

	build, err := c.manager.GetBuildByBuildID(reqBuildUUID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find build %s", reqBuildUUID)) {
		return
	}

	c.render.JSON(w, http.StatusOK, NewResponseBuild(build))
}
