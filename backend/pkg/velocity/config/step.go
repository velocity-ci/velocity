package config

import (
	"encoding/json"
	"fmt"

	v3 "github.com/velocity-ci/velocity/backend/pkg/velocity/docker/compose/v3"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
)

type step interface{}

type baseStep struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type stepSetup struct{}

type stepBlueprint struct {
	baseStep
	Name         string `json:"name"`
	IgnoreErrors string `json:"ignoreErrors"`
}

type stepDockerRun struct {
	baseStep
	Image          string                             `json:"image"`
	Command        v3.DockerComposeServiceCommand     `json:"command"`
	Environment    v3.DockerComposeServiceEnvironment `json:"environment"`
	WorkingDir     string                             `json:"workingDir"`
	MountPoint     string                             `json:"mountPoint"`
	IgnoreExitCode bool                               `json:"ignoreExitCode"`
}

type stepDockerPush struct {
	baseStep
	Tags []string `json:"tags"`
}

type stepDockerCompose struct {
	baseStep
	ComposeFile string `json:"composeFile"`
}

type stepDockerBuild struct {
	baseStep
	Dockerfile string   `json:"dockerfile"`
	Context    string   `json:"context"`
	Tags       []string `json:"tags"`
}

func unmarshalStep(rawMessage []byte) (step, error) {
	var m map[string]interface{}
	err := json.Unmarshal(rawMessage, &m)
	if err != nil {
		return nil, err
	}
	var s step
	switch m["type"] {
	// case "setup":
	// 	s =
	case "run":
		s = &stepDockerRun{
			baseStep: baseStep{
				Type: "run",
			},
			Command: []string{},
		}
	case "build":
		s = &stepDockerBuild{
			baseStep: baseStep{
				Type: "build",
			},
		}
	case "compose":
		s = &stepDockerCompose{
			baseStep: baseStep{
				Type: "compose",
			},
		}
	case "push":
		s = &stepDockerPush{
			baseStep: baseStep{
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

func parseStep(cS step) (*v1.Step, error) {
	switch x := cS.(type) {
	case *stepDockerRun:
		return &v1.Step{
			Description: x.Description,
			Impl: &v1.Step_DockerRun{
				&v1.DockerRun{
					Image:          x.Image,
					Command:        x.Command,
					Environment:    x.Environment,
					WorkingDir:     x.WorkingDir,
					MountPoint:     x.MountPoint,
					IgnoreExitCode: x.IgnoreExitCode,
				},
			},
		}, nil
	case *stepDockerBuild:
		return &v1.Step{
			Description: x.Description,
			Impl: &v1.Step_DockerBuild{
				&v1.DockerBuild{
					Dockerfile: x.Dockerfile,
					Context:    x.Context,
					Tags:       x.Tags,
				},
			},
		}, nil
	case *stepDockerPush:
		return &v1.Step{
			Description: x.Description,
			Impl: &v1.Step_DockerPush{
				&v1.DockerPush{
					Tags: x.Tags,
				},
			},
		}, nil
	case *stepDockerCompose:
		return &v1.Step{
			Description: x.Description,
			Impl: &v1.Step_DockerCompose{
				&v1.DockerCompose{
					ComposeFile: x.ComposeFile,
				},
			},
		}, nil
	}
	return nil, fmt.Errorf("could not determine step from %T", cS)
}
