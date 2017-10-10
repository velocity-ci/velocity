package project

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/middlewares"
)

// Controller - Handles Projects.
type Controller struct {
	logger  *log.Logger
	render  *render.Render
	manager *BoltManager
}

// NewController - Returns a new Controller for Projects.
func NewController(
	projectBoltManager *BoltManager,
	projectResolver *Resolver,
) *Controller {
	return &Controller{
		logger:  log.New(os.Stdout, "[controller:project]", log.Lshortfile),
		render:  render.New(),
		manager: projectBoltManager,
	}
}

// Setup - Sets up the routes for Projects.
func (c Controller) Setup(router *mux.Router) {

	router.Handle("/v1/projects", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectsHandler)),
	)).Methods("GET")

	router.Handle("/v1/projects/{id}", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectHandler)),
	)).Methods("GET")

	router.Handle("/v1/projects/{id}", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.deleteProjectHandler)),
	)).Methods("DELETE")

	router.Handle("/v1/projects", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.postProjectsHandler)),
	)).Methods("POST")

	c.logger.Println("Set up Project controller.")
}

func (c Controller) getProjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	p, err := c.manager.FindByID(vars["id"])

	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	c.render.JSON(w, http.StatusOK, p.ToResponseProject())
}

func (c Controller) getProjectsHandler(w http.ResponseWriter, r *http.Request) {

	projects := c.manager.FindAll()

	responseProjects := []*domain.ResponseProject{}
	for _, p := range projects {
		responseProjects = append(responseProjects, p.ToResponseProject())
	}

	c.render.JSON(w, http.StatusOK, responseProjects)
}

func (c Controller) deleteProjectHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	err := c.manager.DeleteById(vars["id"])

	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	c.render.JSON(w, http.StatusNoContent, nil)
}

func (c Controller) postProjectsHandler(w http.ResponseWriter, r *http.Request) {
	// username := middlewares.UsernameFromContext(r.Context())

	boltProject, err := FromRequest(r.Body)
	if err != nil {
		middlewares.HandleRequestError(err, w, c.render)
		return
	}

	c.manager.Save(boltProject)

	c.render.JSON(w, http.StatusCreated, boltProject.ToResponseProject())
}
