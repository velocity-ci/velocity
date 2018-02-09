package builder

import "time"

type Transport interface {
	WriteJSON(interface{}) error
	ReadJSON(interface{}) error
	Close() error
}

const (
	stateReady = "ready"
	stateBusy  = "busy"
	stateError = "error"
)

type Builder struct {
	ID        string
	State     string
	CreatedAt time.Time
	UpdatedAt time.Time

	ws      Transport
	Command *BuilderCtrlMessage
}
