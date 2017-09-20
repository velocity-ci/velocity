package auth

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/velocity-ci/velocity/master/velocity/domain/user"
)

// Controller - Handles authentication
type Controller struct {
	logger  *log.Logger
	render  *render.Render
	manager *user.BoltManager
}

// NewController - returns a new Controller for Authentication.
func NewController(
	controllerLogger *log.Logger,
	renderer *render.Render,
	userBoltManager *user.BoltManager,
) *Controller {
	return &Controller{
		logger:  controllerLogger,
		render:  renderer,
		manager: userBoltManager,
	}
}

// Setup - Sets up the Auth Controller
func (c Controller) Setup(router *mux.Router) {
	router.
		HandleFunc("/v1/auth", c.authHandler).
		Methods("POST")
	c.logger.Println("Set up Auth controller.")
}

func (c Controller) authHandler(w http.ResponseWriter, r *http.Request) {

	requestUser, err := FromRequest(r.Body)
	if err != nil {
		c.render.JSON(w, http.StatusBadRequest, nil)
		return
	}

	boltUser, err := c.manager.FindByUsername(requestUser.Username)
	if err != nil {
		c.render.JSON(w, http.StatusUnauthorized, nil)
		return
	}

	if !boltUser.ValidatePassword(requestUser.Password) {
		c.render.JSON(w, http.StatusUnauthorized, nil)
		return
	}

	token := NewAuthToken(boltUser)
	c.render.JSON(w, http.StatusCreated, token)

}
