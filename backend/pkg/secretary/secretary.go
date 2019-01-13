package secretary

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/velocity-ci/velocity/backend/pkg/git"
	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

const (
	PoolTopic = "secretaries:pool"
)

type Secretary struct {
	run bool

	baseArchitectAddress string
	secret               string

	http *http.Client
	ws   *phoenix.Client
}

func (b *Secretary) Stop() error {
	b.run = false
	return nil
}
func (b *Secretary) Start() {
	velocity.GetLogger().Info("git version", zap.String("version", git.GetVersion()))
	b.baseArchitectAddress = getArchitectAddress()
	b.secret = getSecretarySecret()
	b.http = &http.Client{
		Timeout: time.Second * 10,
	}

	if !waitForService(b.http, fmt.Sprintf("%s/v1/health", b.baseArchitectAddress)) {
		velocity.GetLogger().Fatal("could not connect to architect", zap.String("address", b.baseArchitectAddress))
		b.Stop()
		return
	}

	velocity.GetLogger().Info("connecting to architect", zap.String("address", b.baseArchitectAddress))
	b.connect()
}

func (b *Secretary) connect() {
	wsAddress := strings.Replace(b.baseArchitectAddress, "http", "ws", 1)
	wsAddress = fmt.Sprintf("%s/socket/v1/secretaries/websocket", wsAddress)

	jobs := map[string]func(payload json.RawMessage) (interface{}, error){
		"vlcty_health-check": func(json.RawMessage) (interface{}, error) {
			return "OK", nil
		},
		"vlcty_repo-get-commits": getCommitsEvent,
	}

	eventHandlers := map[string]func(*phoenix.PhoenixMessage) error{}
	for k, f := range jobs {
		eventHandlers[k] = func(m *phoenix.PhoenixMessage) error {
			res, err := f(m.Payload.(json.RawMessage))

			if err != nil {
				b.ws.Socket.Send(&phoenix.PhoenixMessage{
					Event: phoenix.PhxReplyEvent,
					Topic: PoolTopic,
					Ref:   m.Ref,
					Payload: map[string]interface{}{
						"status": "error",
						"errors": []map[string]string{
							map[string]string{
								"message": err.Error(),
							},
						},
					},
				}, false)

				return err
			}

			b.ws.Socket.Send(&phoenix.PhoenixMessage{
				Event:   phoenix.PhxReplyEvent,
				Topic:   PoolTopic,
				Ref:     m.Ref,
				Payload: res,
			}, false)

			return nil
		}
	}
	ws, err := phoenix.NewClient(wsAddress, eventHandlers)

	if err != nil {
		velocity.GetLogger().Error("could not establish websocket connection", zap.Error(err))
		b.Stop()
		return
	}
	velocity.GetLogger().Debug("established websocket connection", zap.String("address", wsAddress))
	b.ws = ws

	err = b.ws.Subscribe(
		PoolTopic,
		b.secret,
	)
	if err != nil {
		velocity.GetLogger().Error("could not subscribe to builder topic", zap.String("topic", PoolTopic), zap.Error(err))
		b.Stop()
		return
	}

	b.ws.Wait(5)
}

func New() velocity.App {
	return &Secretary{run: true}
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

func getSecretarySecret() string {
	secret := os.Getenv("SECRETARY_SECRET")
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

type registeredSecretary struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

type registerSecretaryResponse struct {
	Data registeredSecretary `json:"data"`
}

func connectToArchitect(address string, secret string) *websocket.Conn {
	wsAddress := strings.Replace(address, "http", "ws", 1)
	headers := http.Header{}
	headers.Set("Authorization", secret)
	var dialer *websocket.Dialer
	conn, _, err := dialer.Dial(
		fmt.Sprintf("%s/secretary/ws", wsAddress),
		headers,
	)

	if err != nil {
		h := sha256.New()
		h.Write([]byte(secret))
		velocity.GetLogger().Fatal("could not connect to architect", zap.String("address", address), zap.String("secretSHA256", string(h.Sum(nil))))
	}

	return conn
}
