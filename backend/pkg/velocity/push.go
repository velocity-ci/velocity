package velocity

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type DockerPush struct {
	BaseStep `yaml:",inline"`
	Tags     []string `json:"tags" yaml:"tags"`
}

func (s *DockerPush) UnmarshalYamlInterface(y map[interface{}]interface{}) error {
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

func NewDockerPush() *DockerPush {
	return &DockerPush{
		Tags: []string{},
		BaseStep: BaseStep{
			Type:          "push",
			OutputStreams: []string{"push"},
			Params:        map[string]Parameter{},
		},
	}
}

func (dP DockerPush) GetDetails() string {
	return fmt.Sprintf("tags: %s", dP.Tags)
}

func (dP *DockerPush) Execute(emitter Emitter, tsk *Task) error {
	writer := emitter.GetStreamWriter("push")
	writer.SetStatus(StateRunning)
	writer.Write([]byte(fmt.Sprintf("\n%s\n## %s\n\x1b[0m", infoANSI, dP.Description)))

	cli, _ := client.NewEnvClient()
	ctx := context.Background()

	for _, t := range dP.Tags {
		imageIDProgress = map[string]string{}
		// Determine correct authToken
		authToken := getAuthToken(t, tsk.Docker.Registries)
		reader, err := cli.ImagePush(ctx, t, types.ImagePushOptions{
			All:          true,
			RegistryAuth: authToken,
		})
		if err != nil {
			log.Println(err)
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("\nPush failed: %s", err)))
			return err
		}
		handleOutput(reader, tsk.ResolvedParameters, writer)
		writer.Write([]byte(fmt.Sprintf("\nPushed: %s", t)))
	}

	writer.SetStatus(StateSuccess)
	writer.Write([]byte(fmt.Sprintf("\n%s\n### SUCCESS\x1b[0m", successANSI)))
	return nil

}

func (dP DockerPush) Validate(params map[string]Parameter) error {
	return nil
}

func (dP *DockerPush) SetParams(params map[string]Parameter) error {
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

func (dP *DockerPush) String() string {
	j, _ := json.Marshal(dP)
	return string(j)
}
