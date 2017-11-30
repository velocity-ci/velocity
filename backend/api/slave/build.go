package slave

import (
	"github.com/velocity-ci/velocity/backend/api/domain/build"
)

type CommandMessage struct {
	Command string  `json:"command"`
	Data    Message `json:"data"`
}

type BuildCommand struct {
	Build build.Build `json:"build"`
}

func NewBuildCommand(b *build.Build) *CommandMessage {
	return &CommandMessage{
		Command: "build",
		Data: BuildCommand{
			Build: *b,
		},
	}
}
