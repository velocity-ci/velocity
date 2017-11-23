package slave

import (
	"github.com/velocity-ci/velocity/backend/api/domain/build"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type CommandMessage struct {
	Command string  `json:"command"`
	Data    Message `json:"data"`
}

type BuildCommand struct {
	Build build.Build
}

func NewBuildCommand(p *project.Project, t *velocity.Task, commitHash string, buildId uint64) *CommandMessage {
	return &CommandMessage{
		Command: "build",
		Data: BuildCommand{
			Task:  t,
			Build: velocity.NewBuild(p.ToTaskProject(), commitHash, buildId),
		},
	}
}
