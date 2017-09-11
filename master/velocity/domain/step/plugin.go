package step

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/velocity-ci/velocity/master/velocity/domain"
)

type Plugin struct {
	domain.BaseStep
	Image          string            `json:"image" yaml:"image"`
	DockerInDocker bool              `json:"dind" yaml:"dind"`
	Environment    map[string]string `json:"environment" yaml:"environment"`
}

func (p *Plugin) SetEmitter(e func(string)) {
	p.Emit = e
}

func (p Plugin) GetType() string {
	return "plugin"
}

func (p Plugin) GetDescription() string {
	return p.Description
}

func (p Plugin) GetDetails() string {
	return fmt.Sprintf("image: %s dind: %v", p.Image, p.DockerInDocker)
}

func (p *Plugin) Execute() error {

	ctx := context.Background()

	stdIn, stdOut, stdErr := term.StdStreams()
	dockerCli := command.NewDockerCli(stdIn, stdOut, stdErr)
	dockerCli.Initialize(cliflags.NewClientOptions())

	res, _ := dockerCli.Client().ImagePull(ctx, fmt.Sprintf("docker.io/%s", p.Image), types.ImagePullOptions{})
	defer res.Close()

	aux := func(auxJSON *json.RawMessage) {
		var result types.BuildResult
		if err := json.Unmarshal(*auxJSON, &result); err != nil {
			fmt.Fprintf(dockerCli.Err(), "Failed to parse aux message: %s", err)
		}
	}

	jsonmessage.DisplayJSONMessagesStream(res, dockerCli.Out(), dockerCli.Out().FD(), false, aux)
	env := []string{}
	for k, v := range p.Environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	config := &container.Config{
		Image: p.Image,
		Cmd:   []string{"/app/run.sh"},
		Env:   env,
	}

	hostConfig := &container.HostConfig{}

	if p.DockerInDocker {
		config.Volumes = map[string]struct{}{
			"/var/run/docker.sock": struct{}{},
		}
		hostConfig.Binds = []string{
			"/var/run/docker.sock:/var/run/docker.sock",
		}
	}

	createResponse, _ := dockerCli.Client().ContainerCreate(ctx, config, hostConfig, nil, "")

	dockerCli.Client().ContainerStart(ctx, createResponse.ID, types.ContainerStartOptions{})

	responseBody, _ := dockerCli.Client().ContainerLogs(ctx, createResponse.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})

	scanner := bufio.NewScanner(responseBody)
	for scanner.Scan() {
		allBytes := scanner.Bytes()
		o := string(allBytes[8:])
		for _, p := range p.Parameters {
			if p.Secret {
				o = strings.Replace(o, p.Value, "***", -1)
			}
		}
		p.Emit(o)
	}

	c, _ := dockerCli.Client().ContainerInspect(ctx, createResponse.ID)
	fmt.Println(c.State.ExitCode)

	return nil
}

func (p Plugin) Validate(params []domain.Parameter) error {

	return nil
}

func (p *Plugin) SetParams(params []domain.Parameter) error {

	return nil
}
