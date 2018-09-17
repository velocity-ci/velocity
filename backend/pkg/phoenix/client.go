package phoenix

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

type wsMessage struct {
	sent     time.Time
	message  *PhoenixMessage
	response chan *PhoenixReplyPayload
}

type Client struct {
	subscribedTopics map[string]PhoenixGuardianJoinPayload

	customEvents map[string]func(*PhoenixMessage) error

	Socket *Socket

	address string
	headers http.Header
}

func NewClient(address string, customEvents map[string]func(*PhoenixMessage) error) (*Client, error) {
	ws := &Client{
		address:          address,
		subscribedTopics: map[string]PhoenixGuardianJoinPayload{},
		customEvents:     customEvents,
	}

	err := ws.dial()
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (c *Client) Wait(reconnects int) {
	for i := 0; i <= reconnects-1; i++ {
		for c.Socket.connected {
			time.Sleep(1 * time.Second)
		}
		c.dial()
	}
}

func (c *Client) dial() error {
	var dialer *websocket.Dialer
	conn, _, err := dialer.Dial(
		c.address,
		c.headers,
	)
	if err != nil {
		return err
	}
	c.Socket = NewSocket(conn, c.customEvents, true)

	for topic, payload := range c.subscribedTopics {
		err := c.Subscribe(topic, payload.Token)
		if err != nil {
			velocity.GetLogger().Warn("could not resubscribe", zap.String("topic", topic))
		}
	}
	return nil
}

func (c *Client) Subscribe(topic, token string) error {
	velocity.GetLogger().Debug("subscribing to", zap.String("topic", topic), zap.Int("token(len)", len(token)))
	resp := c.Socket.Send(&PhoenixMessage{
		Event: PhxJoinEvent,
		Topic: topic,
		Payload: PhoenixGuardianJoinPayload{
			Token: token,
		},
	}, true)

	if resp.Status != ResponseOK {
		velocity.GetLogger().Error("could not subscribe", zap.String("topic", topic), zap.Int("token(len)", len(token)))
		return fmt.Errorf("%v", resp.Response)
	}

	velocity.GetLogger().Debug("subscribed to", zap.String("topic", topic))

	c.subscribedTopics[topic] = PhoenixGuardianJoinPayload{
		Token: token,
	}

	return nil
}
