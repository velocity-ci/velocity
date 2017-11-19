package branch

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/backend/api/auth"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
)

// Controller - Handles Branches.
type Controller struct {
	logger         *log.Logger
	render         *render.Render
	manager        Repository
	projectManager project.Repository
}

// NewController - Returns a new Controller for Branches.
func NewController(
	manager *Manager,
	projectManager *project.Manager,
) *Controller {
	return &Controller{
		logger:         log.New(os.Stdout, "[controller:branch]", log.Lshortfile),
		render:         render.New(),
		manager:        manager,
		projectManager: projectManager,
	}
}

// Setup - Sets up the routes for Branches.
func (c Controller) Setup(router *mux.Router) {
	// GET /v1/projects/{id}/branches
	router.Handle("/v1/projects/{id}/branches", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getBranchesHandler)),
	)).Methods("GET")

	c.logger.Println("Set up Commit controller.")
}

func (c Controller) getBranchesHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["id"]

	p, err := c.projectManager.GetByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	branches, count := c.manager.GetAllByProject(p, Query{})

	respBranches := []*ResponseBranch{}

	for _, b := range branches {
		respBranches = append(respBranches, NewResponseBranch(b))
	}

	c.render.JSON(w, http.StatusOK, ManyResponse{
		Total:  count,
		Result: respBranches,
	})
}
