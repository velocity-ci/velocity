package build

import (
	"fmt"
	"path/filepath"
	"strings"

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

	err = docker.BuildContainer(
		buildContext,
		dB.Dockerfile,
		dB.Tags,
		getSecrets(t.parameters),
		writer,
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
