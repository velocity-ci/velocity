package phoenix

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type wsMessage struct {
	sent     time.Time
	message  *PhoenixMessage
	response chan *PhoenixReplyPayload
}

type Client struct {
	// string:PhoenixGuardianPayload
	subscribedTopics sync.Map

	customEvents map[string]func(*PhoenixMessage) error

	Socket *Socket

	address string
	headers http.Header
}

func NewClient(address string, customEvents map[string]func(*PhoenixMessage) error) (*Client, error) {
	ws := &Client{
		address:      address,
		customEvents: customEvents,
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

	c.subscribedTopics.Range(func(topic, payload interface{}) bool {

		err := c.Subscribe(topic.(string), payload.(PhoenixGuardianJoinPayload).Token)
		if err != nil {
			logging.GetLogger().Warn("could not resubscribe", zap.String("topic", topic.(string)))
		}
		return true
	})

	return nil
}

func (c *Client) Subscribe(topic, token string) error {
	logging.GetLogger().Debug("subscribing to", zap.String("topic", topic), zap.Int("token(len)", len(token)))
	resp := c.Socket.Send(&PhoenixMessage{
		Event: PhxJoinEvent,
		Topic: topic,
		Payload: PhoenixGuardianJoinPayload{
			Token: token,
		},
	}, true)

	if resp.Status != ResponseOK {
		logging.GetLogger().Error("could not subscribe", zap.String("topic", topic), zap.Int("token(len)", len(token)))
		return fmt.Errorf("%v", resp.Response)
	}

	logging.GetLogger().Debug("subscribed to", zap.String("topic", topic))

	c.subscribedTopics.Store(topic, PhoenixGuardianJoinPayload{
		Token: token,
	})

	return nil
}
