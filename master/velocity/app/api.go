package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/velocity-ci/velocity/master/velocity/domain/auth"
	"github.com/velocity-ci/velocity/master/velocity/domain/commit"
	"github.com/velocity-ci/velocity/master/velocity/domain/knownhost"
	"github.com/velocity-ci/velocity/master/velocity/domain/project"
	"github.com/velocity-ci/velocity/master/velocity/domain/user"
	"github.com/velocity-ci/velocity/master/velocity/persisters"
	"github.com/velocity-ci/velocity/master/velocity/routers"
)

// VelocityAPI - The Velocity API app
type VelocityAPI struct {
	Router  *routers.MuxRouter
	server  *http.Server
	bolt    *bolt.DB
	workers sync.WaitGroup
}

// App - For starting and stopping gracefully.
type App interface {
	Start()
	Stop()
}

// New - Returns a new Velocity API app
func New() App {
	velocityAPI := &VelocityAPI{}
	velocityAPI.bolt = persisters.GetBoltDB()

	knownhostManager := knownhost.NewManager()

	userBoltManager := user.NewBoltManager(velocityAPI.bolt)
	projectBoltManager := project.NewBoltManager(velocityAPI.bolt)
	commitBoltManager := commit.NewBoltManager(velocityAPI.bolt)
	knownhostBoltManager := knownhost.NewBoltManager(velocityAPI.bolt, knownhostManager)

	knownhostController := knownhost.NewController(knownhostBoltManager)
	authController := auth.NewController(userBoltManager)
	projectController := project.NewController(projectBoltManager)
	commitController := commit.NewController(commitBoltManager, projectBoltManager)

	velocityAPI.Router = routers.NewMuxRouter([]routers.Routable{
		authController,
		projectController,
		commitController,
		knownhostController,
	}, true)

	port := os.Getenv("PORT")
	velocityAPI.server = &http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		Handler:        velocityAPI.Router.Negroni,
		ReadTimeout:    1 * time.Hour,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return velocityAPI
}

// Stop - Stops the API
func (v *VelocityAPI) Stop() {
	if err := v.server.Shutdown(nil); err != nil {
		panic(err)
	}
	if err := v.bolt.Close(); err != nil {
		panic(err)
	}
}

// Start - Starts the API
func (v *VelocityAPI) Start() {
	if err := v.server.ListenAndServe(); err != nil {
		log.Printf("error: %v", err)
	}
}
