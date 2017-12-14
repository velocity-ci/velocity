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
				c.buildManager.UpdateBuild(buildCommand.Build)
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
			c.buildManager.CreateStreamLine(streamLine)

			if buildStep.Status == velocity.StateWaiting {
				buildStep.Status = lM.Status
				buildStep.StartedAt = time.Now()
				c.buildManager.UpdateBuildStep(buildStep)
			}
			if b.Status == velocity.StateWaiting || b.StartedAt.IsZero() {
				b.Status = lM.Status
				b.StartedAt = time.Now()
				c.buildManager.UpdateBuild(b)
			}

			if lM.Status == velocity.StateSuccess {
				buildStep.Status = velocity.StateSuccess
				buildStep.CompletedAt = time.Now()
				c.buildManager.UpdateBuildStep(buildStep)
				_, total := c.buildManager.GetBuildStepsByBuildID(b.ID) // TODO: cache?
				if buildStep.Number == total-1 {
					b.Status = velocity.StateSuccess
					b.CompletedAt = time.Now()
					c.buildManager.UpdateBuild(b)
				}
			} else if lM.Status == velocity.StateFailed {
				buildStep.Status = velocity.StateFailed
				buildStep.CompletedAt = time.Now()
				c.buildManager.UpdateBuildStep(buildStep)
				b.Status = velocity.StateFailed
				b.CompletedAt = time.Now()
				c.buildManager.UpdateBuild(b)
			}

		}
	}
}
