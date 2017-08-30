package auth

import (
	"log"
	"net/http"

	"github.com/VJftw/velocity/master/velocity/domain/user"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// Controller - Handles authentication
type Controller struct {
	logger        *log.Logger
	render        *render.Render
	userDBManager *user.DBManager
}

// NewController - returns a new Controller for Authentication.
func NewController(
	controllerLogger *log.Logger,
	renderer *render.Render,
	userDBManager *user.DBManager,
) *Controller {
	return &Controller{
		logger:        controllerLogger,
		render:        renderer,
		userDBManager: userDBManager,
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

	user, err := c.userDBManager.FindByUsername(requestUser.Username)
	if err != nil {
		c.render.JSON(w, http.StatusUnauthorized, nil)
		return
	}

	if !user.ValidatePassword(requestUser.Password) {
		c.render.JSON(w, http.StatusUnauthorized, nil)
		return
	}

	token := NewAuthToken(user)
	c.render.JSON(w, http.StatusCreated, token)

}
