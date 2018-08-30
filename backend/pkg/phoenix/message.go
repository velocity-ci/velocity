package phoenix

import (
	"encoding/json"

	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

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
	case PhxLeaveEvent:
	case PhxCloseEvent:
		velocity.GetLogger().Debug("websocket heartbeat")
		break
	default:
		velocity.GetLogger().Warn("no payload found for event", zap.String("event", m.Event))
	}

	return nil
}
