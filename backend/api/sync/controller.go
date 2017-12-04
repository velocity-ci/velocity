package sync

import (
	"log"
	"net/http"
	"os"

	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
	"github.com/velocity-ci/velocity/backend/api/websocket"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/backend/api/auth"
)

// Controller - Handles Syncing.
type Controller struct {
	logger           *log.Logger
	render           *render.Render
	projectManager   project.Repository
	commitManager    commit.Repository
	taskManager      task.Repository
	websocketManager *websocket.Manager
}

// NewController - Returns a new Controller for Syncing.
func NewController(
	projectManager *project.Manager,
	commitManager *commit.Manager,
	taskManager *task.Manager,
	websocketManager *websocket.Manager,
) *Controller {
	return &Controller{
		logger:           log.New(os.Stdout, "[controller:sync]", log.Lshortfile),
		render:           render.New(),
		projectManager:   projectManager,
		commitManager:    commitManager,
		taskManager:      taskManager,
		websocketManager: websocketManager,
	}
}

// Setup - Sets up the routes for Syncing.
func (c Controller) Setup(router *mux.Router) {
	// POST /v1/projects/{id}/sync
	router.Handle("/v1/projects/{id}/sync", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.syncProjectHandler)),
	)).Methods("POST")

	c.logger.Println("Set up Syncing controller.")
}

func (c Controller) syncProjectHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["id"]

	p, err := c.projectManager.GetByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	if p.Synchronising {
		c.render.JSON(w, http.StatusBadRequest, map[string][]string{
			"project": []string{"Already synchronising."},
		},
		)
		return
	}

	p.Synchronising = true
	c.projectManager.Update(p)

	go sync(p, c.projectManager, c.commitManager, c.taskManager, c.websocketManager)

	c.render.JSON(w, http.StatusCreated, project.NewResponseProject(p))
}
