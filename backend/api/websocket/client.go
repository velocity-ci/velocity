package websocket

import (
	"time"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
)

func NewClient(ws *websocket.Conn) *Client {
	return &Client{
		ID:            uuid.NewV4().String(),
		ws:            ws,
		subscriptions: []string{},
	}
}

type Client struct {
	ID            string
	ws            *websocket.Conn
	subscriptions []string
}

func (c *Client) Subscribe(s string) {
	c.subscriptions = append(c.subscriptions, s)
}

func (c *Client) Unsubscribe(s string) {
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
}

type Message interface{}

type EmitMessage struct {
	Subscription string  `json:"subscription"`
	Data         Message `json:"data"`
}

type ClientMessage struct {
	Type  string `json:"type"` //subscribe/unsubscribe
	Route string `json:"route"`
}

type BuildMessage struct {
	Step   uint64     `json:"step"`
	Status string     `json:"status"`
	Log    LogMessage `json:"log"`
}

type LogMessage struct {
	Timestamp time.Time `json:"timestamp"`
	Output    string    `json:"output"`
}
