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
			OutputStreams: []string{"build"},
			Params:        map[string]Parameter{},
		},
	}
}

func (s *DockerBuild) UnmarshalYamlInterface(y map[interface{}]interface{}) error {
	switch x := y["dockerfile"].(type) {
	case interface{}:
		s.Dockerfile = x.(string)
		break
	}

	switch x := y["context"].(type) {
	case interface{}:
		s.Context = x.(string)
		break
	}

	s.Tags = []string{}
	switch x := y["tags"].(type) {
	case []interface{}:
		for _, p := range x {
			s.Tags = append(s.Tags, p.(string))
		}
		break
	}

	return nil
}

func (dB DockerBuild) GetDetails() string {
	return fmt.Sprintf("dockerfile: %s, context: %s, tags: %s", dB.Dockerfile, dB.Context, dB.Tags)
}

func (dB *DockerBuild) Execute(emitter Emitter, t *Task) error {
	writer := emitter.GetStreamWriter("build")
	writer.SetStatus(StateRunning)
	writer.Write([]byte(fmt.Sprintf("\n%s\n## %s\n\x1b[0m", infoANSI, dB.Description)))

	authConfigs := getAuthConfigsMap(t.Docker.Registries)

	err := buildContainer(
		dB.Context,
		dB.Dockerfile,
		dB.Tags,
		t.ResolvedParameters,
		writer,
		authConfigs,
	)

	if err != nil {
		writer.SetStatus(StateFailed)
		writer.Write([]byte(fmt.Sprintf("\n%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
		return err
	}

	writer.SetStatus(StateSuccess)
	writer.Write([]byte(fmt.Sprintf("%s\n### SUCCESS \x1b[0m", successANSI)))
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
