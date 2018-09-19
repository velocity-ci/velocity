package builder

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

const (
	EventStartBuild    = "build:start"
	EventSetKnownHosts = "knownhosts:set"
)

type Builder struct {
	run bool

	baseArchitectAddress string
	secret               string

	id    string
	token string

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
		velocity.GetLogger().Fatal("could not connect to architect", zap.String("address", b.baseArchitectAddress))
	}

	if len(b.id) < 1 {
		err := b.registerWithArchitect()
		if err != nil {
			velocity.GetLogger().Error("could not register builder", zap.Error(err))
			return
		}
	}

	velocity.GetLogger().Info("connecting to architect", zap.String("address", b.baseArchitectAddress), zap.String("builderID", b.id))
	b.connect()
}

func (b *Builder) registerWithArchitect() error {
	address := fmt.Sprintf("%s/v1/builders", b.baseArchitectAddress)
	body, err := json.Marshal(&registerBuilderRequest{Secret: b.secret})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", address, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var respBuilder registerBuilderResponse
	err = decoder.Decode(&respBuilder)
	if err != nil {
		return err
	}

	b.id = respBuilder.ID
	b.token = respBuilder.Token

	velocity.GetLogger().Info("registered builder", zap.String("id", b.id))

	return nil
}

func (b *Builder) connect() {
	wsAddress := strings.Replace(b.baseArchitectAddress, "http", "ws", 1)
	wsAddress = fmt.Sprintf("%s/v1/builders/ws", wsAddress)

	ws, err := phoenix.NewClient(wsAddress, map[string]func(*phoenix.PhoenixMessage) error{
		EventStartBuild: func(m *phoenix.PhoenixMessage) error {
			p := BuildPayload{}
			err := json.Unmarshal(m.Payload.(json.RawMessage), &p)
			if err != nil {
				return err
			}
			b.ws.Socket.ReplyOK(m)
			b.runBuild(&p)
			return nil
		},
		EventSetKnownHosts: func(m *phoenix.PhoenixMessage) error {
			p := KnownHostPayload{}
			err := json.Unmarshal(m.Payload.(json.RawMessage), &p)
			if err != nil {
				return err
			}
			b.updateKnownHosts(&p)
			b.ws.Socket.ReplyOK(m)
			return nil
		},
	})
	if err != nil {
		velocity.GetLogger().Error("could not establish websocket connection", zap.Error(err))
		b.run = false
		return
	}
	velocity.GetLogger().Debug("established websocket connection", zap.String("address", wsAddress))
	b.ws = ws

	topic := fmt.Sprintf("builder:%s", b.id)
	err = b.ws.Subscribe(
		topic,
		b.token,
	)
	if err != nil {
		velocity.GetLogger().Error("could not subscribe to builder topic", zap.String("topic", topic), zap.Error(err))
		b.run = false
		return
	}

	b.ws.Wait(5)
}

func New() velocity.App {
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

type registerBuilderRequest struct {
	Secret string `json:"secret"`
}

type registerBuilderResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
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
