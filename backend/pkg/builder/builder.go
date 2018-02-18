package builder

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"

	"github.com/velocity-ci/velocity/backend/pkg/architect"
	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type Builder struct {
}

func (b *Builder) Start() {
	address := getArchitectAddress()
	secret := getBuilderSecret()
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	if !waitForService(client, address) {
		logrus.Fatalf("Could not connect to: %s", address)
	}

	ws := connectToArchitect(address, secret)

	logrus.Infof("connected to %s", address)

	go monitorCommands(ws)
}

func (b *Builder) Stop() error {
	return nil
}

func New() architect.App {
	velocity.SetLogLevel()
	return &Builder{}
}

func getArchitectAddress() string {
	address := os.Getenv("ARCHITECT_ADDRESS") // http://architect || https://architect
	if address == "" {
		logrus.Fatal("Missing ARCHITECT_ADDRESS environment variable")
	}

	if address[:5] != "https" {
		logrus.Info("WARNING: Builds are not protected by TLS.")
	}

	return address
}

func getBuilderSecret() string {
	secret := os.Getenv("BUILDER_SECRET")
	if secret == "" {
		logrus.Fatal("Missing BUILDER_SECRET environment variable")
	}

	return secret
}

func waitForService(client *http.Client, address string) bool {

	for i := 0; i < 6; i++ {
		logrus.Infof("attempting connection to %s", address)
		_, err := client.Get(address)
		if err != nil {
			logrus.Debugf("connection error: %v", err)
		} else {
			logrus.Infof("%s is alive!", address)
			return true
		}
		time.Sleep(5 * time.Second)
	}

	return false
}

func connectToArchitect(address string, secret string) *websocket.Conn {
	wsAddress := strings.Replace(address, "http", "ws", 1)
	headers := http.Header{}
	headers.Set("Authorization", secret)
	var dialer *websocket.Dialer
	conn, _, err := dialer.Dial(
		fmt.Sprintf("%s/builder/ws", wsAddress),
		headers,
	)

	if err != nil {
		logrus.Fatal(err)
	}

	return conn
}

func monitorCommands(ws *websocket.Conn) {
	for {
		command := &builder.BuilderCtrlMessage{}
		err := ws.ReadJSON(command)
		if err != nil {
			log.Println(err)
			log.Println("Closing WebSocket")
			ws.Close()
			return
		}

		if command.Command == "build" {
			logrus.Infof("Got Build: %v", command.Payload)
			runBuild(command.Payload.(*builder.BuildCtrl), ws)
		} else if command.Command == "known-hosts" {
			logrus.Infof("Got known hosts: %v", command.Payload)
			updateKnownHosts(command.Payload.(*builder.KnownHostCtrl))
		}
	}
}