package knownhost

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/backend/api/auth"
	"github.com/velocity-ci/velocity/backend/api/middleware"
)

// Controller - Handles authentication
type Controller struct {
	logger   *log.Logger
	render   *render.Render
	manager  *Manager
	resolver *Resolver
}

// NewController - returns a new Controller for Authentication.
func NewController(
	manager *Manager,
	resolver *Resolver,
) *Controller {
	return &Controller{
		logger:   log.New(os.Stdout, "[controller:knownhost]", log.Lshortfile),
		render:   render.New(),
		manager:  manager,
		resolver: resolver,
	}
}

// Setup - Sets up the KnownHosts Controller
func (c Controller) Setup(router *mux.Router) {

	router.
		Handle("/v1/ssh/known-hosts", negroni.New(
			auth.NewJWT(c.render),
			negroni.Wrap(http.HandlerFunc(c.postKnownHostsHandler)),
		)).Methods("POST")

	router.
		Handle("/v1/ssh/known-hosts", negroni.New(
			auth.NewJWT(c.render),
			negroni.Wrap(http.HandlerFunc(c.getKnownHostsHandler)),
		)).Methods("GET")

	c.logger.Println("Set up Known Hosts controller.")
}

func (c Controller) postKnownHostsHandler(w http.ResponseWriter, r *http.Request) {

	knownHost, err := c.resolver.FromRequest(r.Body)
	if err != nil {
		middleware.HandleRequestError(err, w, c.render)
		return
	}

	err = c.manager.Save(knownHost)
	if err != nil {
		fmt.Println(err)
		c.render.JSON(w, http.StatusInternalServerError, nil)
		return
	}

	c.render.JSON(w, http.StatusCreated, NewResponseKnownHost(knownHost))
}

func (c Controller) getKnownHostsHandler(w http.ResponseWriter, r *http.Request) {
	knownHosts := c.manager.FindAll()

	responseKnownHosts := []*ResponseKnownHost{}
	for _, k := range knownHosts {
		responseKnownHosts = append(responseKnownHosts, NewResponseKnownHost(&k))
	}

	c.render.JSON(w, http.StatusOK, responseKnownHosts)
}
