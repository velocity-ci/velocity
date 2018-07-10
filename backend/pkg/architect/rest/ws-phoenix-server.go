package rest

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

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

func (c *Client) Subscribe(s string, ref uint64, payload *PhoenixGuardianJoinPayload) {
	_, err := jwt.ParseWithClaims(payload.Token, jwtStandardClaims, jwtKeyFunc)
	if err != nil {
		c.WriteJSON(PhoenixMessage{
			Event: PhxReplyEvent,
			Topic: s,
			Ref:   ref,
			Payload: PhoenixReplyPayload{
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
	c.WriteJSON(PhoenixMessage{
		Event: PhxReplyEvent,
		Topic: s,
		Ref:   ref,
		Payload: PhoenixReplyPayload{
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
	c.WriteJSON(PhoenixMessage{
		Event: PhxReplyEvent,
		Topic: s,
		Ref:   ref,
		Payload: PhoenixReplyPayload{
			Status:   "ok",
			Response: map[string]string{},
		},
	})
}

func (c *Client) HandleHeartbeat(ref uint64) {
	c.WriteJSON(PhoenixMessage{
		Event: PhxReplyEvent,
		Topic: PhxSystemTopic,
		Ref:   ref,
		Payload: PhoenixReplyPayload{
			Status:   "ok",
			Response: map[string]string{},
		},
	})
}

func (c *Client) HandleMessage(m *PhoenixMessage) {
	if m.Topic == PhxSystemTopic {
		switch m.Event {
		case PhxHeartbeatEvent:
			c.HandleHeartbeat(m.Ref)
			break
		}
		return
	}

	switch m.Event {
	case PhxJoinEvent:
		c.Subscribe(m.Topic, m.Ref, m.Payload.(*PhoenixGuardianJoinPayload))
		break
	case PhxLeaveEvent:
		c.Unsubscribe(m.Topic, m.Ref)
		break
	}
}

// Channel Event constants
const (
	PhxCloseEvent     = "phx_close"
	PhxErrorEvent     = "phx_error"
	PhxJoinEvent      = "phx_join"
	PhxReplyEvent     = "phx_reply"
	PhxLeaveEvent     = "phx_leave"
	PhxHeartbeatEvent = "heartbeat"
	PhxSystemTopic    = "phoenix"
)

type PhoenixMessage struct {
	Event   string      `json:"event"`
	Topic   string      `json:"topic"`
	Payload interface{} `json:"payload"`
	Ref     uint64      `json:"ref"`
}

type PhoenixReplyPayload struct {
	Status   string      `json:"status"`
	Response interface{} `json:"response"`
}

type PhoenixGuardianJoinPayload struct {
	Token string `json:"token"`
}

func (m *PhoenixMessage) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	// We'll store the error (if any) so we can return it if necessary
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	// Deserialize Event
	err = json.Unmarshal(*objMap["event"], &m.Event)
	if err != nil {
		return err
	}
	// Deserialize Topic
	err = json.Unmarshal(*objMap["topic"], &m.Topic)
	if err != nil {
		return err
	}
	// Deserialize Ref
	err = json.Unmarshal(*objMap["ref"], &m.Ref)
	if err != nil {
		return err
	}

	// Deserialize Payload by Event
	var rawData json.RawMessage
	err = json.Unmarshal(*objMap["payload"], &rawData)
	if err != nil {
		return err
	}

	switch m.Event {
	case PhxJoinEvent:
		p := PhoenixGuardianJoinPayload{}
		err := json.Unmarshal(rawData, &p)
		if err != nil {
			return err
		}
		m.Payload = &p
		break
	case PhxHeartbeatEvent:
		velocity.GetLogger().Debug("websocket heartbeat")
		break
	default:
		velocity.GetLogger().Warn("no payload found for event", zap.String("event", m.Event))
	}

	return nil
}
