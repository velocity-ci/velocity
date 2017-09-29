package knownhost

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/middlewares"
)

// Controller - Handles authentication
type Controller struct {
	logger  *log.Logger
	render  *render.Render
	manager *BoltManager
}

// NewController - returns a new Controller for Authentication.
func NewController(
	controllerLogger *log.Logger,
	renderer *render.Render,
	manager *BoltManager,
) *Controller {
	return &Controller{
		logger:  controllerLogger,
		render:  renderer,
		manager: manager,
	}
}

// Setup - Sets up the KnownHosts Controller
func (c Controller) Setup(router *mux.Router) {

	router.
		Handle("/v1/ssh/known-hosts", negroni.New(
			middlewares.NewJWT(c.render),
			negroni.Wrap(http.HandlerFunc(c.postKnownHostsHandler)),
		)).Methods("POST")

	router.
		Handle("/v1/ssh/known-hosts", negroni.New(
			middlewares.NewJWT(c.render),
			negroni.Wrap(http.HandlerFunc(c.getKnownHostsHandler)),
		)).Methods("GET")

	c.logger.Println("Set up Known Hosts controller.")
}

func (c Controller) postKnownHostsHandler(w http.ResponseWriter, r *http.Request) {

	boltKnownHost, err := FromRequest(r.Body)
	if err != nil {
		middlewares.HandleRequestError(err, w, c.render)
		return
	}

	err = c.manager.Save(boltKnownHost)
	if err != nil {
		fmt.Println(err)
		c.render.JSON(w, http.StatusInternalServerError, nil)
		return
	}

	c.render.JSON(w, http.StatusCreated, boltKnownHost.ToResponseKnownHost())
}

func (c Controller) getKnownHostsHandler(w http.ResponseWriter, r *http.Request) {
	knownHosts := c.manager.FindAll()

	responseKnownHosts := []*domain.ResponseKnownHost{}
	for _, k := range knownHosts {
		responseKnownHosts = append(responseKnownHosts, k.ToResponseKnownHost())
	}

	c.render.JSON(w, http.StatusOK, responseKnownHosts)
}
