package knownhost

import (
	"fmt"
	"log"
	"net/http"

	ut "github.com/go-playground/universal-translator"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/middlewares"
	validator "gopkg.in/go-playground/validator.v9"
)

// Controller - Handles authentication
type Controller struct {
	logger     *log.Logger
	render     *render.Render
	validator  *validator.Validate
	translator ut.Translator
	manager    *BoltManager
}

// NewController - returns a new Controller for Authentication.
func NewController(
	controllerLogger *log.Logger,
	renderer *render.Render,
	validator *validator.Validate,
	translator ut.Translator,
	manager *BoltManager,
) *Controller {
	return &Controller{
		logger:     controllerLogger,
		render:     renderer,
		validator:  validator,
		translator: translator,
		manager:    manager,
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

	boltKnownHost, err := FromRequest(r.Body, c.validator, c.translator)

	if err != nil {
		if _, ok := err.(validator.ValidationErrors); !ok {
			c.render.JSON(w, http.StatusBadRequest, "Invalid payload.")
			return
		}

		c.render.JSON(w, http.StatusBadRequest, err.(validator.ValidationErrors).Translate(c.translator))
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
