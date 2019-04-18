package build

import (
	"fmt"
	"path/filepath"
	"strings"

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
	return fmt.Sprintf("dockerfile: %s, context: %s, tags: %s", dB.Dockerfile, dB.Context, dB.Tags)
}

func (dB *StepDockerBuild) Execute(emitter Emitter, t *Task) error {
	writer, err := dB.GetStreamWriter(emitter, "build")
	if err != nil {
		return err
	}
	defer writer.Close()
	writer.SetStatus(StateRunning)
	fmt.Fprintf(writer, docker.ColorFmt(docker.ANSIInfo, "-> %s"), dB.Description)

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
		fmt.Fprintf(writer, docker.ColorFmt(docker.ANSIError, "-> failed: %s"), err)

		return err
	}

	writer.SetStatus(StateSuccess)
	fmt.Fprintf(writer, docker.ColorFmt(docker.ANSISuccess, "-> success"))

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
