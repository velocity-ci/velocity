package step

import (
	"fmt"

	"github.com/velocity-ci/velocity/master/velocity/domain"
)

type DockerBuild struct {
	domain.BaseStep `yaml:",inline"`
	Dockerfile      string   `json:"dockerfile" yaml:"dockerfile"`
	Context         string   `json:"context" yaml:"context"`
	Tags            []string `json:"tags" yaml:"tags"`
}

func (dB *DockerBuild) SetEmitter(e func(string)) {
	dB.Emit = e
}

func (dB DockerBuild) GetType() string {
	return "build"
}

func (dB DockerBuild) GetDescription() string {
	return dB.Description
}

func (dB DockerBuild) GetDetails() string {
	return fmt.Sprintf("dockerfile: %s, context: %s, tags: %s", dB.Dockerfile, dB.Context, dB.Tags)
}

func (dB *DockerBuild) Execute() error {
	dB.Emit(fmt.Sprintf("%s\n## %s\n\x1b[0m", infoANSI, dB.Description))

	return buildContainer(
		dB.Context,
		dB.Dockerfile,
		dB.Tags,
		dB.Parameters,
		dB.Emit,
	)
}

func (dB *DockerBuild) Validate(params []domain.Parameter) error {
	return nil
}

func (dB *DockerBuild) SetParams(params []domain.Parameter) error {
	return nil
}
