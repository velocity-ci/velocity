package websocket

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/backend/api/auth"
)

// Controller - Handles Websockets
type Controller struct {
	logger  *log.Logger
	render  *render.Render
	manager *Manager
}

// NewController - returns a new Controller for client Websockets.
func NewController(websocketManager *Manager) *Controller {
	return &Controller{
		logger:  log.New(os.Stdout, "[controller:websocket]", log.Lshortfile),
		render:  render.New(),
		manager: websocketManager,
	}
}

// Setup - Sets up the Websocket Controller
func (c Controller) Setup(router *mux.Router) {

	// /v1/ws
	router.Handle("/v1/ws", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.wsClientHandler)),
	)).Methods("GET")

	c.logger.Println("Set up Websocket controller.")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c Controller) wsClientHandler(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	wsClient := NewClient(ws)

	c.manager.Save(wsClient)

	// Monitor for Messages
	go c.monitor(wsClient)
}

func (c *Controller) monitor(client *Client) {
	for {
		message := &PhoenixMessage{}
		err := client.ws.ReadJSON(message)
		if err != nil {
			log.Println(err)
			log.Println("Closing Client WebSocket")
			client.ws.Close()
			c.manager.Remove(client)
			return
		}
		client.HandleMessage(message)

		// if message.Type == "subscribe" {
		// 	client.Subscribe(message.Route)
		// 	if strings.Contains(message.Route, "builds/") {
		// 		routeParts := strings.Split(message.Route, "/")
		// 		projectID := routeParts[1]
		// 		commitHash := routeParts[3]
		// 		buildID := routeParts[5]

		// 		buildIDU, err := strconv.ParseUint(buildID, 10, 64)
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 		// build := c.commitManager.GetBuild(projectID, commitHash, buildIDU)
		// 		// for stepNumber, sL := range build.StepLogs {
		// 		// 	for _, ls := range sL.Logs {
		// 		// 		for _, l := range ls {
		// 		// 			client.ws.WriteJSON(
		// 		// 				&EmitMessage{
		// 		// 					Subscription: fmt.Sprintf("project/%s/commits/%s/builds/%d", projectID, commitHash, buildIDU),
		// 		// 					Data: BuildMessage{
		// 		// 						Step:   uint64(stepNumber),
		// 		// 						Status: sL.Status,
		// 		// 						Log: LogMessage{
		// 		// 							Timestamp: l.Timestamp,
		// 		// 							Output:    l.Output,
		// 		// 						},
		// 		// 					},
		// 		// 				},
		// 		// 			)
		// 		// 		}
		// 		// 	}
		// 		// }
		// 	}
		// } else if message.Type == "unsubscribe" {
		// 	client.Unsubscribe(message.Route)
		// }
	}
}
