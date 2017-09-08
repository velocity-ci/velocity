package project

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/master/velocity/middlewares"
)

// Controller - Handles Projects.
type Controller struct {
	logger           *log.Logger
	render           *render.Render
	projectDBManager *DBManager
}

// NewController - Returns a new Controller for Projects.
func NewController(
	controllerLogger *log.Logger,
	renderer *render.Render,
	projectDBManager *DBManager,
) *Controller {
	return &Controller{
		logger:           controllerLogger,
		render:           renderer,
		projectDBManager: projectDBManager,
	}
}

// Setup - Sets up the routes for Projects.
func (c Controller) Setup(router *mux.Router) {
	router.Handle("/v1/projects", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectsHandler)),
	)).Methods("GET")

	router.Handle("/v1/projects", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.postProjectsHandler)),
	)).Methods("POST")
	c.logger.Println("Set up Project controller.")
}

func (c Controller) getProjectsHandler(w http.ResponseWriter, r *http.Request) {

	projects := c.projectDBManager.FindAll()

	c.render.JSON(w, http.StatusOK, projects)
}

func (c Controller) postProjectsHandler(w http.ResponseWriter, r *http.Request) {
	// username := middlewares.UsernameFromContext(r.Context())

	project, err := FromRequest(r.Body)
	if err != nil {
		c.render.JSON(w, http.StatusBadRequest, nil)
		return
	}

	_, err = c.projectDBManager.FindByID(project.ID)
	if err == nil {
		c.render.JSON(w, http.StatusBadRequest, nil)
		return
	}

	c.projectDBManager.Save(project)

	c.render.JSON(w, http.StatusCreated, project)
}
