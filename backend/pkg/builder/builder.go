package builder

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/velocity-ci/velocity/backend/pkg/architect"
	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type Builder struct {
	run bool
}

func (b *Builder) Start() {
	address := getArchitectAddress()
	secret := getBuilderSecret()
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	for b.run {
		if !waitForService(client, address) {
			velocity.GetLogger().Fatal("could not connect to architect", zap.String("address", address))
		}

		ws := connectToArchitect(address, secret)

		velocity.GetLogger().Info("connected to architect", zap.String("address", address))

		monitorCommands(ws)
	}
}

func (b *Builder) Stop() error {
	b.run = false
	return nil
}

func New() architect.App {
	return &Builder{run: true}
}

func getArchitectAddress() string {
	address := os.Getenv("ARCHITECT_ADDRESS") // http://architect || https://architect
	if address == "" {
		velocity.GetLogger().Fatal("missing environment variable", zap.String("environment variable", "ARCHITECT_ADDRESS"))
	}

	if address[:5] != "https" {
		velocity.GetLogger().Warn("builds are not protected by TLS")

	}

	return address
}

func getBuilderSecret() string {
	secret := os.Getenv("BUILDER_SECRET")
	if secret == "" {
		velocity.GetLogger().Fatal("missing environment variable", zap.String("environment variable", "ARCHITECT_ADDRESS"))
	}

	return secret
}

func waitForService(client *http.Client, address string) bool {

	for i := 0; i < 6; i++ {
		velocity.GetLogger().Debug("attempting connection to", zap.String("address", address))
		_, err := client.Get(address)
		if err != nil {
			velocity.GetLogger().Debug("connection error", zap.Error(err))
		} else {
			velocity.GetLogger().Debug("connection success")
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
		h := sha256.New()
		h.Write([]byte(secret))
		velocity.GetLogger().Fatal("could not connect to architect", zap.String("address", address), zap.String("secretSHA256", string(h.Sum(nil))))
	}

	return conn
}

func monitorCommands(ws *websocket.Conn) {
	for {
		command := &builder.BuilderCtrlMessage{}
		err := ws.ReadJSON(command)
		if err != nil {
			velocity.GetLogger().Error("could not read websocket message", zap.Error(err))
			ws.Close()
			return
		}

		if command.Command == builder.CommandBuild {
			velocity.GetLogger().Info("got build", zap.Any("payload", command.Payload))
			runBuild(command.Payload.(*builder.BuildCtrl), ws)
		} else if command.Command == builder.CommandKnownHosts {
			velocity.GetLogger().Info("got known hosts", zap.Any("payload", command.Payload))
			updateKnownHosts(command.Payload.(*builder.KnownHostCtrl))
		}
	}
}
