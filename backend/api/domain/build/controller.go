package build

import (
	"log"
	"net/http"
	"os"

	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/task"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
)

// Controller - Handles Builds.
type Controller struct {
	logger         *log.Logger
	render         *render.Render
	manager        Repository
	projectManager project.Repository
	commitManager  commit.Repository
	taskManager    task.Repository
	resolver       *Resolver
}

// NewController - Returns a new Controller for Builds.
func NewController(
	buildResolver *Resolver,
	buildManager Repository,
	projectManager project.Repository,
	commitManager commit.Repository,
	taskManager task.Repository,
) *Controller {
	return &Controller{
		logger:         log.New(os.Stdout, "[controller:build]", log.Lshortfile),
		render:         render.New(),
		manager:        buildManager,
		projectManager: projectManager,
		commitManager:  commitManager,
		taskManager:    taskManager,
		resolver:       buildResolver,
	}
}

// Setup - Sets up the routes for Builds.
func (c Controller) Setup(router *mux.Router) {
	c.addBuildRoutes(router)
	c.addStepRoutes(router)
	c.addStreamRoutes(router)

	c.logger.Println("set up controller.")
}

func handleResourceError(r *render.Render, w http.ResponseWriter, err error, message string) bool {
	if err != nil {
		log.Printf(message)
		r.JSON(w, http.StatusNotFound, message)
		return true
	}

	return false
}
