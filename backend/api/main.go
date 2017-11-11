package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/velocity-ci/velocity/backend/api/slave"
	"github.com/velocity-ci/velocity/backend/api/websocket"
	"github.com/velocity-ci/velocity/backend/velocity"

	"github.com/boltdb/bolt"
	"github.com/velocity-ci/velocity/backend/api/auth"
	"github.com/velocity-ci/velocity/backend/api/commit"
	"github.com/velocity-ci/velocity/backend/api/knownhost"
	"github.com/velocity-ci/velocity/backend/api/project"
)

func main() {

	var returnCode = make(chan int)
	var finishUP = make(chan struct{})
	var done = make(chan struct{})

	var gracefulStop = make(chan os.Signal)

	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		// wait for our os signal to stop the app
		// on the graceful stop channel
		sig := <-gracefulStop
		log.Printf("Caught signal: %+s\n", sig)

		// send message on "finish up" channel to tell the app to
		// gracefully shutdown
		finishUP <- struct{}{}

		// wait for word back if we finished or not
		select {
		case <-time.After(30 * time.Second):
			// timeout after 30 seconds waiting for app to finish
			returnCode <- 1
		case <-done:
			// if we got a message on done, we finished, so end app
			returnCode <- 0
		}
	}()

	a := NewVelocity()
	go a.Start()

	fmt.Println("Waiting for finish")
	// wait for finishUP channel write to close the app down
	<-finishUP
	fmt.Println("Stopping Velocity App")
	a.Stop()
	// write to the done channel to signal we are done.
	done <- struct{}{}
	os.Exit(<-returnCode)
}

// VelocityAPI - The Velocity API app
type VelocityAPI struct {
	Router         *MuxRouter
	server         *http.Server
	bolt           *bolt.DB
	workers        sync.WaitGroup
	buildScheduler *slave.BuildScheduler
}

// App - For starting and stopping gracefully.
type App interface {
	Start()
	Stop()
}

// New - Returns a new Velocity API app
func NewVelocity() App {
	velocityAPI := &VelocityAPI{}
	boltLogger := log.New(os.Stdout, "[bolt]", log.Lshortfile)
	velocityAPI.bolt = NewBoltDB(boltLogger, "velocity.db")
	var wg sync.WaitGroup
	velocityAPI.workers = wg

	validate, translator := newValidator()

	// Auth
	authManager := auth.NewManager(velocityAPI.bolt)
	authController := auth.NewController(authManager)

	// Known Host
	knownHostFileManager := knownhost.NewFileManager()
	knownHostManager := knownhost.NewManager(velocityAPI.bolt, knownHostFileManager)
	knownHostValidator := knownhost.NewValidator(validate, translator, knownHostManager)
	knownHostResolver := knownhost.NewResolver(knownHostValidator)
	knownHostController := knownhost.NewController(knownHostManager, knownHostResolver)

	// Project
	projectSyncManager := project.NewSyncManager(velocity.GitClone)
	projectManager := project.NewManager(projectSyncManager, velocityAPI.bolt)
	projectValidator := project.NewValidator(validate, translator, projectManager)
	projectResolver := project.NewResolver(projectValidator)
	projectController := project.NewController(projectManager, projectResolver)

	// Commit
	commitManager := commit.NewManager(velocityAPI.bolt)
	commitResolver := commit.NewResolver(commitManager)
	commitController := commit.NewController(commitManager, projectManager, commitResolver)

	// Client Websocket
	websocketManager := websocket.NewManager()
	websocketController := websocket.NewController(websocketManager, commitManager)

	// Slave
	slaveManager := slave.NewManager()
	slaveController := slave.NewController(slaveManager, commitManager, websocketManager)

	velocityAPI.buildScheduler = slave.NewBuildScheduler(commitManager, slaveManager, projectManager, &velocityAPI.workers)
	go velocityAPI.buildScheduler.Run()

	velocityAPI.Router = NewMuxRouter([]Routable{
		authController,
		knownHostController,
		projectController,
		commitController,
		slaveController,
		websocketController,
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
	log.Println("Stopping Velocity API")
	v.buildScheduler.Stop()
	v.workers.Wait()
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
