package websocket

import "github.com/gorilla/websocket"

func NewClient(ws *websocket.Conn) *Client {
	return &Client{
		ws:            ws,
		subscriptions: []string{},
	}
}

type Client struct {
	ws            *websocket.Conn
	subscriptions []string
}

type EmitMessage struct {
	Subscription string `json:"subscription"`
}

type ClientSubscribe struct {
	Subscription string `json:"subscribe"`
}
