package build

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/docker"
)

type StepDockerBuild struct {
	BaseStep
	Dockerfile string   `json:"dockerfile"`
	Context    string   `json:"context"`
	Tags       []string `json:"tags"`

	builder *docker.ImageBuilder
}

func NewStepDockerBuild(c *v1.Step_DockerBuild) *StepDockerBuild {
	return &StepDockerBuild{
		BaseStep:   newBaseStep("build", []string{"build"}),
		Dockerfile: c.DockerBuild.GetDockerfile(),
		Context:    c.DockerBuild.GetContext(),
		Tags:       c.DockerBuild.GetTags(),
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

	dB.builder = docker.NewImageBuilder()

	err = dB.builder.Build(
		writer,
		getSecrets(t.parameters),
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

func (dB *StepDockerBuild) Stop() error {
	if dB.builder != nil {
		dB.builder.Stop()
	}
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
