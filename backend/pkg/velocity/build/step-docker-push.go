package build

import (
	"fmt"
	"strings"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/docker"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"go.uber.org/zap"
)

type StepDockerPush struct {
	BaseStep
	Tags []string `json:"tags"`
}

func NewStepDockerPush(c *config.StepDockerPush) *StepDockerPush {
	return &StepDockerPush{
		BaseStep: newBaseStep("push", []string{"push"}),
		Tags:     c.Tags,
	}
}

func (dP StepDockerPush) GetDetails() string {
	return fmt.Sprintf("tags: %s", dP.Tags)
}

func (dP *StepDockerPush) Execute(emitter Emitter, tsk *Task) error {
	writer, err := dP.GetStreamWriter(emitter, "push")
	if err != nil {
		return err
	}
	defer writer.Close()
	writer.SetStatus(StateRunning)
	fmt.Fprintf(writer, "\r")

	for _, t := range dP.Tags {
		err := docker.PushImage(
			writer,
			t,
			GetAddressAuthTokensMap(tsk.Docker.Registries),
			getSecrets(tsk.parameters),
		)
		if err != nil {
			logging.GetLogger().Error("could not push docker image", zap.String("image", t), zap.Error(err))
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, docker.ColorFmt(docker.ANSIError, "-> push failed: %s", "\n"), err)
			return err
		}
	}

	writer.SetStatus(StateSuccess)
	fmt.Fprintf(writer, docker.ColorFmt(docker.ANSISuccess, "-> success", "\n"))
	return nil

}

func (dP StepDockerPush) Validate(params map[string]Parameter) error {
	return nil
}

func (dP *StepDockerPush) SetParams(params map[string]*Parameter) error {
	for paramName, param := range params {
		tags := []string{}
		for _, c := range dP.Tags {
			correctedTag := strings.Replace(c, fmt.Sprintf("${%s}", paramName), param.Value, -1)
			tags = append(tags, correctedTag)
		}
		dP.Tags = tags
	}
	return nil
}
