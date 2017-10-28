package slave

import (
	"github.com/velocity-ci/velocity/backend/api/project"
	"github.com/velocity-ci/velocity/backend/task"
)

type CommandMessage struct {
	Command string  `json:"command"`
	Data    Message `json:"data"`
}

type BuildCommand struct {
	Project    *project.Project `json:"project"`
	Task       *task.Task       `json:"task"`
	CommitHash string           `json:"commit"`
	BuildID    uint64           `json:"buildId`
}

func NewBuildCommand(p *project.Project, t *task.Task, commitHash string, buildId uint64) *CommandMessage {
	return &CommandMessage{
		Command: "build",
		Data: BuildCommand{
			Project:    p,
			Task:       t,
			CommitHash: commitHash,
			BuildID:    buildId,
		},
	}
}
