package slave

import (
	"fmt"

	"github.com/docker/go/canonical/json"
	"github.com/velocity-ci/velocity/backend/api/domain/build"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/knownhost"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
)

type CommandMessage struct {
	Command string  `json:"command"`
	Data    Message `json:"data"`
}

func (c CommandMessage) String() string {
	j, _ := json.Marshal(c)
	return string(j)
}

type BuildCommand struct {
	Build      build.Build       `json:"build"`
	Project    project.Project   `json:"project"`
	Commit     commit.Commit     `json:"commit"`
	Task       task.Task         `json:"task"`
	BuildSteps []build.BuildStep `json:"buildSteps"`
}

func (c BuildCommand) String() string {
	j, _ := json.Marshal(c)
	return string(j)
}

type KnownHostCommand struct {
	KnownHosts []knownhost.KnownHost `json:"knownHosts"`
}

type SlaveBuildLogMessage struct {
	BuildStepID string `json:"buildStepId"`
	StreamName  string `json:"streamName"`
	LineNumber  uint64 `json:"lineNumber"`
	Status      string `json:"status"`
	Output      string `json:"output"`
}

func NewBuildCommand(b build.Build, p project.Project, c commit.Commit, task task.Task) CommandMessage {
	return CommandMessage{
		Command: "build",
		Data: BuildCommand{
			Build:   b,
			Project: p,
			Commit:  c,
			Task:    task,
		},
	}
}

func NewKnownHostCommand(knownHosts []knownhost.KnownHost) CommandMessage {
	return CommandMessage{
		Command: "known-hosts",
		Data: KnownHostCommand{
			KnownHosts: knownHosts,
		},
	}
}

func (c *CommandMessage) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	// We'll store the error (if any) so we can return it if necessary
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	// Deserialize Command
	err = json.Unmarshal(*objMap["command"], &c.Command)
	if err != nil {
		return err
	}

	// Deserialize Data by command
	var rawData json.RawMessage
	err = json.Unmarshal(*objMap["data"], &rawData)
	if err != nil {
		return err
	}

	if c.Command == "build" {
		d := BuildCommand{}
		err := json.Unmarshal(rawData, &d)
		if err != nil {
			return err
		}
		c.Data = &d
	} else if c.Command == "known-hosts" {
		d := KnownHostCommand{}
		err := json.Unmarshal(rawData, &d)
		if err != nil {
			return err
		}
		c.Data = &d
	} else {
		return fmt.Errorf("unsupported type in json.Unmarshal: %s", c.Command)
	}

	return nil
}
