package build

import (
	"fmt"
	"strings"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/docker"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
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

func (dP *StepDockerPush) Execute(emitter out.Emitter, tsk *Task) error {
	writer := emitter.GetStreamWriter("push")
	defer writer.Close()
	writer.SetStatus(StateRunning)
	fmt.Fprintf(writer, out.ColorFmt(out.ANSIInfo, "-> %s"), dP.Description)

	for _, t := range dP.Tags {
		err := docker.PushImage(
			writer,
			t,
			GetAddressAuthTokensMap(tsk.Docker.Registries),
			getSecrets(tsk.Parameters),
		)
		if err != nil {
			logging.GetLogger().Error("could not push docker image", zap.String("image", t), zap.Error(err))
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, out.ColorFmt(out.ANSIError, "-> push failed: %s"), err)
			return err
		}
	}

	writer.SetStatus(StateSuccess)
	fmt.Fprintf(writer, out.ColorFmt(out.ANSISuccess, "-> success"))
	return nil

}

func (dP StepDockerPush) Validate(params map[string]Parameter) error {
	return nil
}

func (dP *StepDockerPush) SetParams(params map[string]Parameter) error {
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
