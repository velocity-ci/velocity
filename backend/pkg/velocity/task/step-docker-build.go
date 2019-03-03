package task

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
)

type DockerBuild struct {
	BaseStep
	Dockerfile string   `json:"dockerfile"`
	Context    string   `json:"context"`
	Tags       []string `json:"tags"`
}

func NewDockerBuild() *DockerBuild {
	return &DockerBuild{
		Dockerfile: "",
		Context:    "",
		Tags:       []string{},
		BaseStep:   newBaseStep("build", []string{"build"}),
	}
}

func (dB DockerBuild) GetDetails() string {
	return fmt.Sprintf("dockerfile: %s, context: %s, tags: %s", dB.Dockerfile, dB.Context, dB.Tags)
}

func (dB *DockerBuild) Execute(emitter out.Emitter, t *Task) error {
	writer := emitter.GetStreamWriter("build")
	defer writer.Close()
	writer.SetStatus(StateRunning)
	fmt.Fprintf(writer, out.ColorFmt(out.ANSIInfo, "-> %s"), dB.Description)

	authConfigs := GetAuthConfigsMap(t.Docker.Registries)

	buildContext := filepath.Join(dB.ProjectRoot, dB.Context)

	err := buildContainer(
		buildContext,
		dB.Dockerfile,
		dB.Tags,
		t.ResolvedParameters,
		writer,
		authConfigs,
	)

	if err != nil {
		writer.SetStatus(StateFailed)
		fmt.Fprintf(writer, out.ColorFmt(out.ANSIError, "-> failed: %s"), err)

		return err
	}

	writer.SetStatus(StateSuccess)
	fmt.Fprintf(writer, out.ColorFmt(out.ANSISuccess, "-> success"))

	return nil
}

func (dB *DockerBuild) Validate(params map[string]Parameter) error {
	return nil
}

func (dB *DockerBuild) SetParams(params map[string]Parameter) error {
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
