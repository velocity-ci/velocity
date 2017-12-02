package velocity

import (
	"fmt"
	"strings"
)

type DockerBuild struct {
	BaseStep   `yaml:",inline"`
	Dockerfile string   `json:"dockerfile" yaml:"dockerfile"`
	Context    string   `json:"context" yaml:"context"`
	Tags       []string `json:"tags" yaml:"tags"`
}

func NewDockerBuild() *DockerBuild {
	return &DockerBuild{
		Dockerfile: "",
		Context:    "",
		Tags:       []string{},
		BaseStep: BaseStep{
			Type:          "build",
			OutputStreams: []OutputStream{NewOutputStream("", "build")},
		},
	}
}

func (dB DockerBuild) GetDetails() string {
	return fmt.Sprintf("dockerfile: %s, context: %s, tags: %s", dB.Dockerfile, dB.Context, dB.Tags)
}

func (dB *DockerBuild) Execute(emitter Emitter, params map[string]Parameter) error {
	emitter.Write([]byte(fmt.Sprintf("%s\n## %s\n\x1b[0m", infoANSI, dB.Description)))

	err := buildContainer(
		dB.Context,
		dB.Dockerfile,
		dB.Tags,
		params,
		emitter,
	)

	if err != nil {
		emitter.SetStatus("failed")
		emitter.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
		return err
	}

	emitter.SetStatus("success")
	emitter.Write([]byte(fmt.Sprintf("%s\n### SUCCESS \x1b[0m", successANSI)))
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
