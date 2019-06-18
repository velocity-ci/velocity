package config

import (
	"encoding/json"
	"fmt"

	v3 "github.com/velocity-ci/velocity/backend/pkg/velocity/docker/compose/v3"
)

type Step interface{}

type BaseStep struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type StepSetup struct{}

type StepBlueprint struct {
	BaseStep
	Name         string `json:"name"`
	IgnoreErrors string `json:"ignoreErrors"`
}

type StepDockerRun struct {
	BaseStep
	Image          string                             `json:"image"`
	Command        v3.DockerComposeServiceCommand     `json:"command"`
	Environment    v3.DockerComposeServiceEnvironment `json:"environment"`
	WorkingDir     string                             `json:"workingDir"`
	MountPoint     string                             `json:"mountPoint"`
	IgnoreExitCode bool                               `json:"ignoreExitCode"`
}

type StepDockerPush struct {
	BaseStep
	Tags []string `json:"tags"`
}

type StepDockerCompose struct {
	BaseStep
	ComposeFile string `json:"composeFile"`
}

type StepDockerBuild struct {
	BaseStep
	Dockerfile string   `json:"dockerfile"`
	Context    string   `json:"context"`
	Tags       []string `json:"tags"`
}

func unmarshalStep(rawMessage []byte) (Step, error) {
	var m map[string]interface{}
	err := json.Unmarshal(rawMessage, &m)
	if err != nil {
		return nil, err
	}
	var s Step
	switch m["type"] {
	// case "setup":
	// 	s =
	case "run":
		s = &StepDockerRun{
			BaseStep: BaseStep{
				Type: "run",
			},
			Command: []string{},
		}
	case "build":
		s = &StepDockerBuild{
			BaseStep: BaseStep{
				Type: "build",
			},
		}
	case "compose":
		s = &StepDockerCompose{
			BaseStep: BaseStep{
				Type: "compose",
			},
		}
	case "push":
		s = &StepDockerPush{
			BaseStep: BaseStep{
				Type: "push",
			},
		}
	}

	if s == nil {
		return nil, fmt.Errorf("could not determine step %+v", m)
	}

	err = json.Unmarshal(rawMessage, s)
	return s, err
}
