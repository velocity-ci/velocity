package slave

// slave websocket endpoint
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
	"github.com/velocity-ci/velocity/backend/api/commit"
)

// Controller - Handles Slaves
type Controller struct {
	logger        *log.Logger
	render        *render.Render
	manager       *Manager
	commitManager *commit.Manager
}

// NewController - returns a new Controller for Slaves.
func NewController(
	slaveManager *Manager,
	commitManager *commit.Manager,
) *Controller {
	return &Controller{
		logger:        log.New(os.Stdout, "[controller:slave]", log.Lshortfile),
		render:        render.New(),
		manager:       slaveManager,
		commitManager: commitManager,
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

	s := c.manager.GetSlaveByID(slaveID)
	s.SetWebSocket(ws)
	s.State = "ready"

	c.manager.Save(s)

	// Monitor for Messages
	go c.monitor(s)
}

func (c *Controller) monitor(s *Slave) {
	for {
		message := &SlaveMessage{}
		err := s.ws.ReadJSON(message)
		if err != nil {
			log.Println(err)
			log.Println("Closing Slave WebSocket")
			s.ws.Close()
			s.ws = nil
			s.State = "disconnected"
			if s.Command != nil && s.Command.Command == "build" {
				buildCommand := s.Command.Data.(BuildCommand)
				build := c.commitManager.GetBuild(buildCommand.Project.ID, buildCommand.CommitHash, buildCommand.Build.ID)
				build.Status = "waiting"
				build.StepLogs = []commit.StepLog{}
				c.commitManager.SaveBuild(build, buildCommand.Project.ID, buildCommand.CommitHash)
			}
			c.manager.Save(s)
			return
		}

		if message.Type == "log" {
			lM := message.Data.(*LogMessage)
			log.Println(lM.Step, lM.Status, lM.Output)
			build := c.commitManager.GetBuild(lM.ProjectID, lM.CommitHash, lM.BuildID)

			if lM.Step > 0 && build.Task.Steps[lM.Step-1].GetType() == "compose" {
			} else {
				if build.StepLogs == nil {
					build.StepLogs = []commit.StepLog{commit.StepLog{Logs: map[string][]commit.Log{}, Status: lM.Status}}
				} else if len(build.StepLogs) <= int(lM.Step) {
					build.StepLogs = append(build.StepLogs, commit.StepLog{Logs: map[string][]commit.Log{}, Status: lM.Status})
				}
				build.StepLogs[lM.Step].Status = lM.Status
				build.StepLogs[lM.Step].Logs["container"] = append(build.StepLogs[lM.Step].Logs["container"], commit.Log{
					Timestamp: time.Now(),
					Output:    lM.Output,
				})
			}
			// TODO: Emit to websocket clients

			if lM.Status == "failed" {
				build.Status = "failed"
				c.commitManager.SaveBuild(build, lM.ProjectID, lM.CommitHash)
				c.commitManager.RemoveQueuedBuild(lM.ProjectID, lM.CommitHash, lM.BuildID)
				s.State = "ready"
				s.Command = nil
				c.manager.Save(s)
			} else if lM.Status == "success" {
				if int(lM.Step) == len(build.Task.Steps) {
					// successfully finished build
					build.Status = "success"
					c.commitManager.SaveBuild(build, lM.ProjectID, lM.CommitHash)
					c.commitManager.RemoveQueuedBuild(lM.ProjectID, lM.CommitHash, lM.BuildID)
					s.State = "ready"
					s.Command = nil
					c.manager.Save(s)
				}
			}
			c.commitManager.SaveBuild(build, lM.ProjectID, lM.CommitHash)
		}
	}
}
