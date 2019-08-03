package build

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/docker/docker/client"
	"github.com/ghodss/yaml"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/docker"
)

type StepDockerBuild struct {
	BaseStep
	Dockerfile string   `json:"dockerfile"`
	Context    string   `json:"context"`
	Tags       []string `json:"tags"`

	builder *docker.ImageBuilder
}

func NewStepDockerBuild(c *config.StepDockerBuild) *StepDockerBuild {
	return &StepDockerBuild{
		BaseStep:   newBaseStep("build", []string{"build"}),
		Dockerfile: c.Dockerfile,
		Context:    c.Context,
		Tags:       c.Tags,
	}
}

func (dB StepDockerBuild) GetDetails() string {
	type details struct {
		Dockerfile string   `json:"dockerfile"`
		Context    string   `json:"context"`
		Tags       []string `json:"tags"`
	}
	y, _ := yaml.Marshal(&details{
		Dockerfile: dB.Dockerfile,
		Context:    dB.Context,
		Tags:       dB.Tags,
	})
	return string(y)
}

func (dB *StepDockerBuild) Execute(emitter Emitter, t *Task) error {
	writer, err := dB.GetStreamWriter(emitter, "build")
	if err != nil {
		return err
	}
	defer writer.Close()
	writer.SetStatus(StateBuilding)
	fmt.Fprintf(writer, "\r")

	authConfigs := GetAuthConfigsMap(t.Docker.Registries)

	buildContext := filepath.Join(t.ProjectRoot, dB.Context)

	cli, _ := client.NewEnvClient()
	ctx := context.Background()

	dB.builder = docker.NewImageBuilder(cli, ctx, writer, getSecrets(t.parameters))

	err = dB.builder.Build(
		buildContext,
		dB.Dockerfile,
		dB.Tags,
		authConfigs,
	)

	if err != nil {
		writer.SetStatus(StateFailed)
		fmt.Fprintf(writer, output.ColorFmt(output.ANSIError, "-> failed: %s", "\n"), err)

		return err
	}

	writer.SetStatus(StateSuccess)
	fmt.Fprintf(writer, output.ColorFmt(output.ANSISuccess, "-> success", "\n"))

	return nil
}

func (dB *StepDockerBuild) GracefulStop() error {
	dB.builder.GracefulStop()
	return nil
}

func (dB *StepDockerBuild) Validate(params map[string]Parameter) error {
	return nil
}

func (dB *StepDockerBuild) SetParams(params map[string]*Parameter) error {
	for paramName, param := range params {
		dB.Context = strings.Replace(dB.Context, fmt.Sprintf("${%s}", paramName), param.Value, -1)
		dB.Dockerfile = strings.Replace(dB.Dockerfile, fmt.Sprintf("${%s}", paramName), param.Value, -1)

		tags := []string{}

		for _, t := range dB.Tags {
			tags = append(tags, strings.Replace(t, fmt.Sprintf("${%s}", paramName), param.Value, -1))
		}
		dB.Tags = tags
	}
	return nil
}
