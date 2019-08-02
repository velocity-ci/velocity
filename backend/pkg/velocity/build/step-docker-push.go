package build

import (
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"

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
	type details struct {
		Tags []string `json:"tags"`
	}
	y, _ := yaml.Marshal(&details{
		Tags: dP.Tags,
	})
	return string(y)
}

func (dP *StepDockerPush) Execute(emitter Emitter, tsk *Task) error {
	writer, err := dP.GetStreamWriter(emitter, "push")
	if err != nil {
		return err
	}
	defer writer.Close()
	writer.SetStatus(StateBuilding)
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
			fmt.Fprintf(writer, output.ColorFmt(output.ANSIError, "-> push failed: %s", "\n"), err)
			return err
		}
	}

	writer.SetStatus(StateSuccess)
	fmt.Fprintf(writer, output.ColorFmt(output.ANSISuccess, "-> success", "\n"))
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
