package build

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/backend/api/auth"
)

func (c Controller) addStreamRoutes(router *mux.Router) {
	// GET /v1/steps/{stepUUID}/streams
	router.Handle("/v1/steps/{stepUUID}/streams", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getStepByUUIDStreams)),
	)).Methods("GET")

	// GET /v1/streams/{streamUUID}
	router.Handle("/v1/streams/{streamUUID}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getStreamByUUID)),
	)).Methods("GET")
}

func (c Controller) getStepByUUIDStreams(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqStepUUID := reqVars["stepUUID"]

	buildStep, err := c.manager.GetBuildStepByBuildStepID(reqStepUUID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find step %s", reqStepUUID)) {
		return
	}

	outputStreams, count := c.manager.GetStreamsByBuildStepID(buildStep.ID)

	respOutputStreams := []ResponseOutputStream{}
	for _, outputStream := range outputStreams {
		respOutputStreams = append(respOutputStreams, ResponseOutputStream{
			ID:   outputStream.ID,
			Name: outputStream.Name,
		})
	}

	c.render.JSON(w, http.StatusOK, OutputStreamManyResponse{
		Total:  count,
		Result: respOutputStreams,
	})
}

func (c Controller) getStreamByUUID(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqStreamUUID := reqVars["streamUUID"]

	stream, err := c.manager.GetStreamByID(reqStreamUUID)
	if handleResourceError(c.render, w, err, fmt.Sprintf("could not find stream %s", reqStreamUUID)) {
		return
	}

	opts := StreamLineQueryOptsFromRequest(r)
	log.Println(opts)

	streamLines, total := c.manager.GetStreamLinesByStreamID(stream.ID, opts)

	c.render.JSON(w, http.StatusOK, StreamLineManyResponse{
		Total:  total,
		Result: streamLines,
	})
}
