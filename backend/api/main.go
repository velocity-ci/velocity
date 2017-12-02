package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/velocity-ci/velocity/backend/api/domain/branch"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/api/domain/user"
	"github.com/velocity-ci/velocity/backend/velocity"

	"github.com/boltdb/bolt"
	"github.com/velocity-ci/velocity/backend/api/auth"
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
	Router *MuxRouter
	server *http.Server
	bolt   *bolt.DB
	// workers        sync.WaitGroup
	// buildScheduler *slave.BuildScheduler
}

// App - For starting and stopping gracefully.
type App interface {
	Start()
	Stop()
}

// New - Returns a new Velocity API app
func NewVelocity() App {
	velocityAPI := &VelocityAPI{}
	// boltLogger := log.New(os.Stdout, "[bolt]", log.Lshortfile)
	// velocityAPI.bolt = NewBoltDB(boltLogger, "velocity.db")
	// var wg sync.WaitGroup
	// velocityAPI.workers = wg

	validate, translator := newValidator()

	// Persistence
	gorm := NewGORMDB()

	// User
	userManager := user.NewManager(gorm)

	// Auth
	auth.EnsureAdminUser(userManager)
	authController := auth.NewController(userManager)

	// Known Host
	// knownHostFileManager := knownhost.NewFileManager()
	// knownHostManager := knownhost.NewManager(velocityAPI.bolt, knownHostFileManager)
	// knownHostValidator := knownhost.NewValidator(validate, translator, knownHostManager)
	// knownHostResolver := knownhost.NewResolver(knownHostValidator)
	// knownHostController := knownhost.NewController(knownHostManager, knownHostResolver)

	// Project
	projectManager := project.NewManager(gorm, velocity.GitClone)
	projectValidator := project.NewValidator(validate, translator, projectManager)
	projectResolver := project.NewResolver(projectValidator)
	projectController := project.NewController(projectManager, projectResolver)

	// Commit
	commitManager := commit.NewManager(gorm)
	commitController := commit.NewController(commitManager, projectManager)

	// Branch
	branchManager := branch.NewManager(gorm)
	branchController := branch.NewController(branchManager, projectManager)

	// Task
	// taskManager := task.NewManager(gorm)
	// taskController := task.NewController(taskManager, projectManager, commitManager)

	// Build
	// buildManager := build.NewManager(gorm)
	// buildResolver := build.NewResolver(taskManager)
	// buildController := build.NewController(buildResolver, buildManager, projectManager, commitManager)

	// // Sync
	// syncController := apiSync.NewController(projectManager, commitManager, branchManager, taskManager)

	// // Slave
	// slaveManager := slave.NewManager(buildManager)
	// slaveController := slave.NewController(slaveManager, buildManager, commitManager)

	// Client Websocket
	// websocketManager := websocket.NewManager()
	// websocketController := websocket.NewController(websocketManager, commitManager)

	// velocityAPI.buildScheduler = slave.NewBuildScheduler(slaveManager, buildManager, &velocityAPI.workers)
	// go velocityAPI.buildScheduler.Run()

	velocityAPI.Router = NewMuxRouter([]Routable{
		authController,
		// knownHostController,
		projectController,
		commitController,
		branchController,
		taskController,
		buildController,
		syncController,
		slaveController,
		// websocketController,
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
	// v.buildScheduler.Stop()
	// v.workers.Wait()
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
