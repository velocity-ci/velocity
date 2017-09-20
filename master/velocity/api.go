package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/unrolled/render"
	"github.com/velocity-ci/velocity/master/velocity/domain/auth"
	"github.com/velocity-ci/velocity/master/velocity/domain/project"
	"github.com/velocity-ci/velocity/master/velocity/domain/user"
	"github.com/velocity-ci/velocity/master/velocity/persisters"
	"github.com/velocity-ci/velocity/master/velocity/routers"
)

// VelocityAPI - The Velocity API app
type VelocityAPI struct {
	Router *routers.MuxRouter
	server *http.Server
	bolt   *bolt.DB
}

// NewVelocityAPI - Returns a new Velocity API app
func NewVelocityAPI() App {
	velocityAPI := &VelocityAPI{}

	controllerLogger := log.New(os.Stdout, "[controller]", log.Lshortfile)
	boltLogger := log.New(os.Stdout, "[bolt]", log.Lshortfile)
	renderer := render.New()

	velocityAPI.bolt = persisters.NewBoltDB(boltLogger)

	userBoltManager := user.NewBoltManager(boltLogger, velocityAPI.bolt)
	projectBoltManager := project.NewBoltManager(boltLogger, velocityAPI.bolt)

	authController := auth.NewController(controllerLogger, renderer, userBoltManager)
	projectController := project.NewController(controllerLogger, renderer, projectBoltManager)

	velocityAPI.Router = routers.NewMuxRouter([]routers.Routable{
		authController,
		projectController,
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
