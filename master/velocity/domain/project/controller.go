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

	router.Handle("/v1/projects/{id}", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectHandler)),
	)).Methods("GET")

	router.Handle("/v1/projects", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.postProjectsHandler)),
	)).Methods("POST")

	// GET /v1/projects/{id}/commits
	router.Handle("/v1/projects/{id}/commits", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.syncProjectHandler)),
	)).Methods("GET")

	// POST /v1/projects/{id}/sync
	router.Handle("/v1/projects/{id}/sync", negroni.New(
		middlewares.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.syncProjectHandler)),
	)).Methods("POST")

	c.logger.Println("Set up Project controller.")
}

func (c Controller) getProjectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	p, err := c.projectDBManager.FindByID(vars["id"])

	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	c.render.JSON(w, http.StatusOK, p)
}

func (c Controller) getProjectsHandler(w http.ResponseWriter, r *http.Request) {

	projects := c.projectDBManager.FindAll()

	c.render.JSON(w, http.StatusOK, projects)
}

func (c Controller) postProjectsHandler(w http.ResponseWriter, r *http.Request) {
	// username := middlewares.UsernameFromContext(r.Context())

	p, err := FromRequest(r.Body)
	if err != nil {
		c.render.JSON(w, http.StatusBadRequest, nil)
		return
	}
	valid, errs := ValidatePOST(p, c.projectDBManager)
	if !valid {
		c.render.JSON(w, http.StatusBadRequest, errs)
		return
	}

	c.projectDBManager.Save(p)

	c.render.JSON(w, http.StatusCreated, p)
}

func (c Controller) syncProjectHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["id"]

	p, err := c.projectDBManager.FindByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	// if p.Synchronising {
	// 	c.render.JSON(w, http.StatusBadRequest, nil)
	// 	return
	// }

	p.Synchronising = true
	c.projectDBManager.Save(p)
	go sync(p)
	c.render.JSON(w, http.StatusCreated, p)
}
