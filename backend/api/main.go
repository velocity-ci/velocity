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

	"github.com/velocity-ci/velocity/backend/api/domain/build"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/knownhost"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
	"github.com/velocity-ci/velocity/backend/api/domain/user"
	"github.com/velocity-ci/velocity/backend/api/slave"
	apiSync "github.com/velocity-ci/velocity/backend/api/sync"
	"github.com/velocity-ci/velocity/backend/api/websocket"
	"github.com/velocity-ci/velocity/backend/velocity"

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
	Router  *MuxRouter
	server  *http.Server
	wg      sync.WaitGroup
	workers []Worker
}

// App - For starting and stopping gracefully.
type App interface {
	Start()
	Stop()
}

type Worker interface {
	StartWorker()
	StopWorker()
}

// New - Returns a new Velocity API app
func NewVelocity() App {
	velocityAPI := &VelocityAPI{}

	validate, translator := newValidator()

	// Persistence
	gorm := NewGORMDB()

	// Client Websocket
	websocketManager := websocket.NewManager()
	websocketController := websocket.NewController(websocketManager)

	// User
	userManager := user.NewManager(gorm)

	// Auth
	auth.EnsureAdminUser(userManager)
	authController := auth.NewController(userManager)

	// Known Host
	knownHostManager := knownhost.NewManager(gorm, websocketManager)
	knownHostValidator := knownhost.NewValidator(validate, translator, knownHostManager)
	knownHostResolver := knownhost.NewResolver(knownHostValidator)
	knownHostController := knownhost.NewController(knownHostManager, knownHostResolver)

	// Project
	projectManager := project.NewManager(gorm, velocity.GitClone, websocketManager)
	projectValidator := project.NewValidator(validate, translator, projectManager)
	projectResolver := project.NewResolver(projectValidator)
	projectController := project.NewController(projectManager, projectResolver)

	// Commit
	commitManager := commit.NewManager(gorm, websocketManager)
	commitController := commit.NewController(commitManager, projectManager)

	// Task
	taskManager := task.NewManager(gorm, websocketManager)
	taskController := task.NewController(taskManager, projectManager, commitManager)

	// Build
	fileManager := build.NewFileManager(&velocityAPI.wg)
	velocityAPI.workers = append(velocityAPI.workers, fileManager)
	buildManager := build.NewManager(gorm, fileManager, websocketManager)
	buildResolver := build.NewResolver(commitManager)
	buildController := build.NewController(buildResolver, buildManager, projectManager, commitManager, taskManager)

	// // Sync
	syncController := apiSync.NewController(projectManager, commitManager, taskManager, websocketManager)

	// Slave
	slaveManager := slave.NewManager(buildManager, taskManager, commitManager, projectManager, knownHostManager, websocketManager)
	slaveController := slave.NewController(slaveManager, buildManager, commitManager, websocketManager)

	buildScheduler := slave.NewBuildScheduler(slaveManager, buildManager, &velocityAPI.wg)
	velocityAPI.workers = append(velocityAPI.workers, buildScheduler)

	for _, w := range velocityAPI.workers {
		go w.StartWorker()
	}

	velocityAPI.Router = NewMuxRouter([]Routable{
		authController,
		knownHostController,
		projectController,
		commitController,
		taskController,
		buildController,
		syncController,
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
	for _, w := range v.workers {
		w.StopWorker()
	}
	v.wg.Wait()
	if err := v.server.Shutdown(nil); err != nil {
		panic(err)
	}
}

// Start - Starts the API
func (v *VelocityAPI) Start() {
	if err := v.server.ListenAndServe(); err != nil {
		log.Printf("error: %v", err)
	}
}
