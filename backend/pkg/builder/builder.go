package builder

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

const (
	PoolTopic = "builders:pool"
)

type Builder struct {
	run bool

	baseArchitectAddress string
	secret               string

	//id    string
	//token string

	http *http.Client
	ws   *phoenix.Client
}

func (b *Builder) Stop() error {
	b.run = false
	return nil
}
func (b *Builder) Start() {
	b.baseArchitectAddress = getArchitectAddress()
	b.secret = getBuilderSecret()
	b.http = &http.Client{
		Timeout: time.Second * 10,
	}

	if !waitForService(b.http, fmt.Sprintf("%s/v1/health", b.baseArchitectAddress)) {
		logging.GetLogger().Fatal("could not connect to architect", zap.String("address", b.baseArchitectAddress))
		b.Stop()
		return
	}

	logging.GetLogger().Info("connecting to architect", zap.String("address", b.baseArchitectAddress))
	b.connect()
}

func (b *Builder) connect() {
	wsAddress := strings.Replace(b.baseArchitectAddress, "http", "ws", 1)
	wsAddress = fmt.Sprintf("%s/socket/v1/builders/websocket", wsAddress)

	eventHandlers := map[string]func(*phoenix.PhoenixMessage) error{}
	for _, j := range jobs {
		eventHandlers[fmt.Sprintf("%s%s", EventJobDoPrefix, j.GetName())] = func(m *phoenix.PhoenixMessage) error {
			payloadBytes, _ := json.Marshal(m.Payload)
			err := j.Parse(payloadBytes)
			if err != nil {
				logging.GetLogger().Error("could not unmarshal payload", zap.Error(err))
			}

			err = j.Do(b.ws)
			if err != nil {
				b.ws.Socket.Send(&phoenix.PhoenixMessage{
					Event: EventJobStatus,
					Topic: fmt.Sprintf("job:%s", j.GetID()),
					Payload: map[string]interface{}{
						"status": "error",
						"errors": []map[string]string{
							map[string]string{
								"message": err.Error(),
							},
						},
					},
				}, false)
			}

			b.ws.Socket.Send(&phoenix.PhoenixMessage{
				Event: EventJobStatus,
				Topic: fmt.Sprintf("job:%s", j.GetID()),
				Payload: map[string]interface{}{
					"status": "success",
				},
			}, false)

			SendBuilderReady(b.ws)
			return err
		}
	}
	eventHandlers[EventJobStop] = func(*phoenix.PhoenixMessage) error {
		return nil
	}
	ws, err := phoenix.NewClient(wsAddress, eventHandlers)

	if err != nil {
		logging.GetLogger().Error("could not establish websocket connection", zap.Error(err))
		b.Stop()
		return
	}
	logging.GetLogger().Debug("established websocket connection", zap.String("address", wsAddress))
	b.ws = ws

	err = b.ws.Subscribe(
		PoolTopic,
		b.secret,
	)
	if err != nil {
		logging.GetLogger().Error("could not subscribe to builder topic", zap.String("topic", PoolTopic), zap.Error(err))
		b.Stop()
		return
	}

	SendBuilderReady(b.ws)

	b.ws.Wait(5)
}

func New() velocity.App {
	return &Builder{run: true}
}

func getArchitectAddress() string {
	address := os.Getenv("ARCHITECT_ADDRESS") // http://architect || https://architect
	if address == "" {
		logging.GetLogger().Fatal("missing environment variable", zap.String("environment variable", "ARCHITECT_ADDRESS"))
	}

	if address[:5] != "https" {
		logging.GetLogger().Warn("builds are not protected by TLS")

	}

	return address
}

func getBuilderSecret() string {
	secret := os.Getenv("BUILDER_SECRET")
	if secret == "" {
		logging.GetLogger().Fatal("missing environment variable", zap.String("environment variable", "ARCHITECT_ADDRESS"))
	}

	return secret
}

func waitForService(client *http.Client, address string) bool {

	for i := 0; i < 6; i++ {
		logging.GetLogger().Debug("attempting connection to", zap.String("address", address))
		_, err := client.Get(address)
		if err != nil {
			logging.GetLogger().Debug("connection error", zap.Error(err))
		} else {
			logging.GetLogger().Debug("connection success")
			return true
		}
		time.Sleep(5 * time.Second)
	}

	return false
}

type registerBuilderRequest struct {
	Secret string `json:"secret"`
}

type registeredBuilder struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type registerBuilderResponse struct {
	Data registeredBuilder `json:"data"`
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
		logging.GetLogger().Fatal("could not connect to architect", zap.String("address", address), zap.String("secretSHA256", string(h.Sum(nil))))
	}

	return conn
}
