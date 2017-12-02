package slave

import (
	"github.com/docker/go/canonical/json"
	"github.com/velocity-ci/velocity/backend/api/domain/build"
)

type CommandMessage struct {
	Command string  `json:"command"`
	Data    Message `json:"data"`
}

type BuildCommand struct {
	Build      build.Build       `json:"build"`
	BuildSteps []build.BuildStep `json:"buildSteps"`
}

func (c BuildCommand) String() string {
	j, _ := json.Marshal(c)
	return string(j)
}

type SlaveBuildLogMessage struct {
	OutputStreamID string `json:"outputStreamID"`
	LineNumber     uint64 `json:"lineNumber"`
	Output         string `json:"output"`
	Status         string `json:"status"`
}

func NewBuildCommand(b build.Build, bS []build.BuildStep) CommandMessage {
	return CommandMessage{
		Command: "build",
		Data: BuildCommand{
			Build:      b,
			BuildSteps: bS,
		},
	}
}
