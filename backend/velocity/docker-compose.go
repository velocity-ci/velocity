package velocity

import (
	"fmt"
)

type DockerCompose struct {
	BaseStep
	ComposeFile string `json:"composeFile" yaml:"compose_file"`
	Contents    dockerComposeYaml
}

func (dC DockerCompose) GetType() string {
	return "compose"
}

func (dC DockerCompose) GetDescription() string {
	return dC.Description
}

func (dC DockerCompose) GetDetails() string {
	return fmt.Sprintf("composeFile: %s", dC.ComposeFile)
}

func (dC *DockerCompose) Validate(params []Parameter) error {
	return nil
}

func (dC *DockerCompose) SetParams(params []Parameter) error {
	return nil
}

func (dC *DockerCompose) Execute(emitter Emitter, params map[string]Parameter) error {

	return nil
}

type dockerComposeYaml struct {
	Services map[string]dockerComposeService `json:"services" yaml:"services"`
}

type dockerComposeService struct {
	Build       dockerComposeServiceBuild `json:"build" yaml:"build"`
	WorkingDir  string                    `json:"workingDir" yaml:"working_dir"`
	Command     string                    `json:"command" yaml:"command"`
	Links       []string                  `json:"links" yaml:"links"`
	Environment map[string]string         `json:"environment" yaml:"environment"`
	Volumes     []string                  `json:"volumes" yaml:"volumes"`
	Expose      []string                  `json:"expose" yaml:"expose"`
}

type dockerComposeServiceBuild struct {
	Context    string `json:"context" yaml:"context"`
	Dockerfile string `json:"dockerfile" yaml:"dockerfile"`
}
