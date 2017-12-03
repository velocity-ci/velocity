package slave

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/backend/api/auth"
	"github.com/velocity-ci/velocity/backend/api/domain/build"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	apiWebsocket "github.com/velocity-ci/velocity/backend/api/websocket"
	"github.com/velocity-ci/velocity/backend/velocity"
)

// Controller - Handles Slaves
type Controller struct {
	logger           *log.Logger
	render           *render.Render
	manager          *Manager
	buildManager     build.Repository
	commitManager    *commit.Manager
	websocketManager *apiWebsocket.Manager
}

// NewController - returns a new Controller for Slaves.
func NewController(
	slaveManager *Manager,
	buildManager *build.Manager,
	commitManager *commit.Manager,
	websocketManager *apiWebsocket.Manager,
) *Controller {
	return &Controller{
		logger:           log.New(os.Stdout, "[controller:slave]", log.Lshortfile),
		render:           render.New(),
		manager:          slaveManager,
		commitManager:    commitManager,
		buildManager:     buildManager,
		websocketManager: websocketManager,
	}
}

// Setup - Sets up the Auth Controller
func (c Controller) Setup(router *mux.Router) {

	router.
		HandleFunc("/v1/slaves", c.postSlavesHandler).
		Methods("POST")

	// /v1/slaves/ws
	router.Handle("/v1/slaves/ws", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.wsSlavesHandler)),
	)).Methods("GET")

	c.logger.Println("Set up Slave controller.")
}

func (c Controller) postSlavesHandler(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Authorization") != fmt.Sprintf("basic %s", os.Getenv("SLAVE_SECRET")) {
		c.render.JSON(w, http.StatusUnauthorized, nil)
		return
	}

	reqSlave, err := FromRequest(r.Body)
	if err != nil {
		c.render.JSON(w, http.StatusBadRequest, nil)
		return
	}

	// TODO: Passkey authentication middleware

	// TODO: Validation (unique ID)

	s := NewSlave(reqSlave.ID)
	c.manager.Save(s)

	token := auth.NewAuthToken(s.ID)

	c.render.JSON(w, http.StatusCreated, token)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (c Controller) wsSlavesHandler(w http.ResponseWriter, r *http.Request) {

	slaveID := auth.UsernameFromContext(r.Context())

	if c.manager.WebSocketConnected(slaveID) {
		c.render.JSON(w, http.StatusBadRequest, "Slave already connected")
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	s, _ := c.manager.GetSlaveByID(slaveID)
	s.SetWebSocket(ws)
	s.State = "ready"

	c.manager.Save(s)

	// Monitor for Messages
	go c.monitor(s)
}

func (c *Controller) monitor(s Slave) {
	for {
		message := &SlaveMessage{}
		err := s.ws.ReadJSON(message)
		if err != nil {
			log.Println(err)
			log.Println("Closing Slave WebSocket")
			s.ws.Close()
			s.ws = nil
			s.State = "disconnected"
			log.Println(s.Command)
			if s.Command.Command == "build" {
				buildCommand := s.Command.Data.(BuildCommand)
				buildCommand.Build.Status = "waiting"
				c.buildManager.SaveBuild(buildCommand.Build)
			}
			c.manager.Save(s)
			return
		}

		if message.Type == "log" {
			lM := message.Data.(*SlaveBuildLogMessage)
			buildStep, err := c.buildManager.GetBuildStepByBuildStepID(lM.BuildStepID)
			if err != nil {
				log.Printf("could not find buildStep %s", lM.BuildStepID)
				return
			}
			b, err := c.buildManager.GetBuildByBuildID(buildStep.BuildID)
			if err != nil {
				log.Printf("could not find build %s", buildStep.BuildID)
			}
			outputStream, err := c.buildManager.GetStreamByBuildStepIDAndStreamName(buildStep.ID, lM.StreamName) // TODO: Cache in memory
			if err != nil {
				log.Printf("could not find buildStep:stream %s:%s", lM.BuildStepID, lM.StreamName)
				return
			}

			streamLine := build.NewStreamLine(outputStream.ID, lM.LineNumber, time.Now(), lM.Output)
			c.buildManager.SaveStreamLine(streamLine)
			c.websocketManager.EmitAll(&apiWebsocket.PhoenixMessage{
				Topic:   fmt.Sprintf("stream:%s", streamLine.BuildStepStreamID),
				Event:   "new",
				Payload: streamLine,
			})
			updateBuildStep := false
			updateBuild := false
			if buildStep.Status == velocity.StateWaiting {
				buildStep.Status = lM.Status
				buildStep.StartedAt = time.Now()
				updateBuildStep = true
			}
			if b.Status == velocity.StateWaiting || b.StartedAt.IsZero() {
				b.Status = lM.Status
				b.StartedAt = time.Now()
				updateBuild = true
			}

			if lM.Status == velocity.StateSuccess {
				buildStep.Status = velocity.StateSuccess
				buildStep.CompletedAt = time.Now()
				updateBuildStep = true
				_, total := c.buildManager.GetBuildStepsByBuildID(b.ID) // TODO: cache?
				if buildStep.Number == total-1 {
					b.Status = velocity.StateSuccess
					b.CompletedAt = time.Now()
					updateBuild = true
				}
			} else if lM.Status == velocity.StateFailed {
				buildStep.Status = velocity.StateFailed
				buildStep.CompletedAt = time.Now()
				updateBuildStep = true
				b.Status = velocity.StateFailed
				b.CompletedAt = time.Now()
				updateBuild = true
			}

			if updateBuildStep {
				c.buildManager.SaveBuildStep(buildStep)
				c.websocketManager.EmitAll(&apiWebsocket.PhoenixMessage{
					Topic:   fmt.Sprintf("step:%s", buildStep.ID),
					Event:   "modify",
					Payload: buildStep,
				})

			}

			if updateBuild {
				c.buildManager.SaveBuild(b)
				c.websocketManager.EmitAll(&apiWebsocket.PhoenixMessage{
					Topic:   fmt.Sprintf("build:%s", b.ID),
					Event:   "modify",
					Payload: b,
				})
			}

			// topics
			// stream: uuid
			// step: uuid
			// build: uuid

			// OLD ---
			// log.Println(lM.Step, lM.Status, lM.Output)
			// build := c.commitManager.GetBuild(lM.ProjectID, lM.CommitHash, lM.BuildID)
			// timestamp := time.Now()

			// if lM.Step > 0 && build.Task.Steps[lM.Step-1].GetType() == "compose" {
			// } else {
			// 	if build.StepLogs == nil {
			// 		build.StepLogs = []commit.StepLog{commit.StepLog{Logs: map[string][]commit.Log{}, Status: lM.Status}}
			// 	} else if len(build.StepLogs) <= int(lM.Step) {
			// 		build.StepLogs = append(build.StepLogs, commit.StepLog{Logs: map[string][]commit.Log{}, Status: lM.Status})
			// 	}
			// 	build.StepLogs[lM.Step].Status = lM.Status
			// 	build.StepLogs[lM.Step].Logs["container"] = append(build.StepLogs[lM.Step].Logs["container"], commit.Log{
			// 		Timestamp: timestamp,
			// 		Output:    lM.Output,
			// 	})
			// }

			// if lM.Status == "failed" {
			// 	build.Status = "failed"
			// 	c.commitManager.SaveBuild(build, lM.ProjectID, lM.CommitHash)
			// 	c.commitManager.RemoveQueuedBuild(lM.ProjectID, lM.CommitHash, lM.BuildID)
			// 	s.State = "ready"
			// 	s.Command = nil
			// 	c.manager.Save(s)
			// } else if lM.Status == "success" {
			// 	if int(lM.Step) == len(build.Task.Steps)-1 {
			// 		// successfully finished build
			// 		build.Status = "success"
			// 		c.commitManager.SaveBuild(build, lM.ProjectID, lM.CommitHash)
			// 		c.commitManager.RemoveQueuedBuild(lM.ProjectID, lM.CommitHash, lM.BuildID)
			// 		s.State = "ready"
			// 		s.Command = nil
			// 		c.manager.Save(s)
			// 	}
			// }
			// c.commitManager.SaveBuild(build, lM.ProjectID, lM.CommitHash)

			// // Emit to websocket clients
			// c.websocketManager.EmitAll(
			// 	&apiWebsocket.EmitMessage{
			// 		Subscription: fmt.Sprintf("project/%s/commits/%s/builds/%d", lM.ProjectID, lM.CommitHash, lM.BuildID),
			// 		Data: apiWebsocket.BuildMessage{
			// 			Step:   lM.Step,
			// 			Status: lM.Status,
			// 			Log: apiWebsocket.LogMessage{
			// 				Timestamp: timestamp,
			// 				Output:    lM.Output,
			// 			},
			// 		},
			// 	},
			// )
		}
	}
}
