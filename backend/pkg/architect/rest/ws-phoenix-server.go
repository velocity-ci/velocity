package rest

import (
	"fmt"
	"os"
	"sync"

	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

// NewClient - returns a new websocket client mirroring the Phoenix Framework's (http://phoenixframework.org/) websocket protocol
func NewClient(ws *websocket.Conn) *Client {
	return &Client{
		ID:            uuid.NewV4().String(),
		ws:            ws,
		subscriptions: []string{},
	}
}

type Client struct {
	ID            string
	alive         bool
	ws            *websocket.Conn
	wsLock        sync.RWMutex
	subscriptions []string
}

func jwtKeyFunc(t *jwt.Token) (interface{}, error) {
	// Check the signing method (from echo.labstack.jwt middleware)
	if t.Method.Alg() != jwtSigningMethod.Name {
		return nil, fmt.Errorf("Unexpected jwt signing method=%v", t.Header["alg"])
	}
	return []byte(os.Getenv("JWT_SECRET")), nil
}

func (c *Client) WriteJSON(v interface{}) error {
	c.wsLock.Lock()
	defer c.wsLock.Unlock()
	return c.ws.WriteJSON(v)
}

func (c *Client) Subscribe(s string, ref uint64, payload *phoenix.PhoenixGuardianJoinPayload) {
	_, err := jwt.ParseWithClaims(payload.Token, jwtStandardClaims, jwtKeyFunc)
	if err != nil {
		c.WriteJSON(phoenix.PhoenixMessage{
			Event: phoenix.PhxReplyEvent,
			Topic: s,
			Ref:   ref,
			Payload: phoenix.PhoenixReplyPayload{
				Status: "error",
				Response: map[string]string{
					"message": "access denied",
				},
			},
		})
		velocity.GetLogger().Warn("could not authenticate client to channel", zap.String("clientID", c.ID), zap.Error(err))
		return
	}

	c.subscriptions = append(c.subscriptions, s)
	c.WriteJSON(phoenix.PhoenixMessage{
		Event: phoenix.PhxReplyEvent,
		Topic: s,
		Ref:   ref,
		Payload: phoenix.PhoenixReplyPayload{
			Status:   "ok",
			Response: map[string]string{},
		},
	})
}

func (c *Client) Unsubscribe(s string, ref uint64) {
	var element int
	for i, v := range c.subscriptions {
		if v == s {
			element = i
			break
		}
	}
	if element < len(c.subscriptions) {
		c.subscriptions = append(c.subscriptions[:element], c.subscriptions[element+1:]...)
	}
	c.WriteJSON(phoenix.PhoenixMessage{
		Event: phoenix.PhxReplyEvent,
		Topic: s,
		Ref:   ref,
		Payload: phoenix.PhoenixReplyPayload{
			Status:   phoenix.ResponseOK,
			Response: map[string]string{},
		},
	})
}

func (c *Client) HandleHeartbeat(ref uint64) {
	c.WriteJSON(phoenix.PhoenixMessage{
		Event: phoenix.PhxReplyEvent,
		Topic: phoenix.PhxSystemTopic,
		Ref:   ref,
		Payload: phoenix.PhoenixReplyPayload{
			Status:   phoenix.ResponseOK,
			Response: map[string]string{},
		},
	})
}

func (c *Client) HandleMessage(m *phoenix.PhoenixMessage) {
	if m.Topic == phoenix.PhxSystemTopic {
		switch m.Event {
		case phoenix.PhxHeartbeatEvent:
			c.HandleHeartbeat(m.Ref)
			break
		}
		return
	}

	switch m.Event {
	case phoenix.PhxJoinEvent:
		c.Subscribe(m.Topic, m.Ref, m.Payload.(*phoenix.PhoenixGuardianJoinPayload))
		break
	case phoenix.PhxLeaveEvent:
		c.Unsubscribe(m.Topic, m.Ref)
		break
	}
}
