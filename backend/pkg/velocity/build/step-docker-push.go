package build

import (
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/docker"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"go.uber.org/zap"
)

type StepDockerPush struct {
	BaseStep
	Tags []string `json:"tags"`

	pusher *docker.ImagePusher
}

func NewStepDockerPush(c *v1.Step_DockerPush) *StepDockerPush {
	return &StepDockerPush{
		BaseStep: newBaseStep("push", []string{"push"}),
		Tags:     c.DockerPush.GetTags(),
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

	dP.pusher = docker.NewImagePusher()

	for _, t := range dP.Tags {
		err := dP.pusher.Push(
			writer,
			getSecrets(tsk.parameters),
			t,
			GetAddressAuthTokensMap(tsk.Docker.Registries),
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

func (dP *StepDockerPush) Stop() error {
	if dP.pusher != nil {
		dP.pusher.Stop()
	}
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
