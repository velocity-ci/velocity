package builder

import (
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
)

const (
	StateReady        = "ready"
	StateBusy         = "busy"
	StateError        = "error"
	StateDisconnected = "disconnected"
)

type Builder struct {
	ID        string
	Token     string
	State     string
	CreatedAt time.Time
	UpdatedAt time.Time

	WS *phoenix.Server
}
