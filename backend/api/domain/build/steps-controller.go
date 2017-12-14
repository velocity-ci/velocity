package build

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/backend/api/auth"
)

func (c Controller) addStepRoutes(router *mux.Router) {

	// GET /v1/builds/{buildUUID}/steps
	router.Handle("/v1/builds/{buildUUID}/steps", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getBuildByUUIDStepsHandler)),
	)).Methods("GET")

	// GET /v1/steps/{stepUUID}
	router.Handle("/v1/steps/{stepUUID}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getStepByUUID)),
	)).Methods("GET")
}

func (c Controller) getBuildByUUIDStepsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqBuildUUID := reqVars["buildUUID"]

	build, err := c.manager.GetBuildByBuildID(reqBuildUUID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find build %s", reqBuildUUID)) {
		return
	}

	t, err := c.taskManager.GetByTaskID(build.TaskID)
	if err != nil {
		handleResourceError(c.render, w, err, fmt.Sprintf("could not find task %s for build %s?!?!", build.TaskID, build.ID))
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

func (c Controller) getStepByUUID(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqStepUUID := reqVars["stepUUID"]

	buildStep, err := c.manager.GetBuildStepByBuildStepID(reqStepUUID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find step %s", reqStepUUID)) {
		return
	}

	build, err := c.manager.GetBuildByBuildID(buildStep.BuildID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find build %s for build step %s?!?!", buildStep.BuildID, buildStep.ID)) {
		return
	}

	t, err := c.taskManager.GetByTaskID(build.TaskID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find task %s for build %s?!?!", build.TaskID, build.ID)) {
		return
	}

	c.render.JSON(w, http.StatusOK, NewResponseBuildStep(buildStep, t.Steps[buildStep.Number]))
}
