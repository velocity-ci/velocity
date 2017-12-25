package slave

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

type RequestSlave struct {
	ID string `json:"id"`
}

type ResponseSlave struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

func NewResponseSlave(s Slave) ResponseSlave {
	return ResponseSlave{
		ID:    s.ID,
		State: s.State,
	}
}

type ManyResponse struct {
	Total  uint64          `json:"total"`
	Result []ResponseSlave `json:"result"`
}

type Slave struct {
	ID      string
	State   string // ready, busy, disconnected
	ws      *websocket.Conn
	Command CommandMessage
}

type SlaveQuery struct {
	Amount uint64
	Page   uint64
	Status string
}

func NewSlave(ID string) Slave {
	return Slave{
		ID:    ID,
		State: "disconnected",
	}
}

func (s *Slave) SetWebSocket(ws *websocket.Conn) {
	s.ws = ws
}

type SlaveMessage struct {
	Type string  `json:"type"`
	Data Message `json:"data"`
}

type Message interface{}

type SlaveStreamLine struct {
	OutputStreamID string `json:"outputStreamId"`
	Status         string `json:"status"`
	LineNumber     uint64 `json:"lineNumber"`
	Output         string `json:"output"`
}

func (c *SlaveMessage) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	// We'll store the error (if any) so we can return it if necessary
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	// Deserialize Command
	err = json.Unmarshal(*objMap["type"], &c.Type)
	if err != nil {
		return err
	}

	// Deserialize Data by command
	var rawData json.RawMessage
	err = json.Unmarshal(*objMap["data"], &rawData)
	if err != nil {
		return err
	}

	if c.Type == "log" {
		d := SlaveBuildLogMessage{}
		err := json.Unmarshal(rawData, &d)
		if err != nil {
			return err
		}
		c.Data = &d
	} else {
		return fmt.Errorf("unsupported type in json.Unmarshal: %s", c.Type)
	}

	return nil
}
