package step

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/velocity-ci/velocity/master/velocity/domain"
)

type DockerRun struct {
	domain.BaseStep
	Image       string            `json:"image" yaml:"image"`
	Command     []string          `json:"command" yaml:"command"`
	Environment map[string]string `json:"environment" yaml:"environment"`
	WorkingDir  string            `json:"workingDir" yaml:"working_dir"`
	MountPoint  string            `json:"mountPoint" yaml:"mount_point"`
}

func (dB *DockerRun) SetEmitter(e func(string)) {
	dB.Emit = e
}

func (dB DockerRun) GetType() string {
	return "run"
}

func (dB DockerRun) GetDescription() string {
	return dB.Description
}

func (dB DockerRun) GetDetails() string {
	return fmt.Sprintf("image: %s command: %s", dB.Image, dB.Command)
}

func (dB *DockerRun) Execute() error {
	ctx := context.Background()

	stdIn, stdOut, stdErr := term.StdStreams()
	dockerCli := command.NewDockerCli(stdIn, stdOut, stdErr)
	dockerCli.Initialize(cliflags.NewClientOptions())

	res, _ := dockerCli.Client().ImagePull(ctx, fmt.Sprintf("docker.io/%s", dB.Image), types.ImagePullOptions{})
	defer res.Close()

	aux := func(auxJSON *json.RawMessage) {
		var result types.BuildResult
		if err := json.Unmarshal(*auxJSON, &result); err != nil {
			fmt.Fprintf(dockerCli.Err(), "Failed to parse aux message: %s", err)
		}
	}

	jsonmessage.DisplayJSONMessagesStream(res, dockerCli.Out(), dockerCli.Out().FD(), false, aux)
	if dB.MountPoint == "" {
		dB.MountPoint = "/velocity_ci"
	}
	env := []string{}
	for k, v := range dB.Environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cwd, _ := os.Getwd()
	createResponse, _ := dockerCli.Client().ContainerCreate(ctx, &container.Config{
		Image: dB.Image,
		Cmd:   dB.Command,
		Volumes: map[string]struct{}{
			dB.MountPoint: struct{}{},
		},
		WorkingDir: fmt.Sprintf("%s/%s", dB.MountPoint, dB.WorkingDir),
		Env:        env,
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/%s", cwd, dB.MountPoint),
		},
	}, nil, "")

	dockerCli.Client().ContainerStart(ctx, createResponse.ID, types.ContainerStartOptions{})

	responseBody, _ := dockerCli.Client().ContainerLogs(ctx, createResponse.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})

	scanner := bufio.NewScanner(responseBody)
	for scanner.Scan() {
		allBytes := scanner.Bytes()
		o := string(allBytes[8:])
		for _, p := range dB.Parameters {
			if p.Secret {
				o = strings.Replace(o, p.Value, "***", -1)
			}
		}
		dB.Emit(o)
	}

	c, _ := dockerCli.Client().ContainerInspect(ctx, createResponse.ID)
	fmt.Println(c.State.ExitCode)

	return nil
}

func (dR DockerRun) Validate(params []domain.Parameter) error {
	re := regexp.MustCompile("\\$\\{(.+)\\}")

	requiredParams := re.FindAllStringSubmatch(dR.Image, -1)
	if !isAllInParams(requiredParams, params) {
		return fmt.Errorf("Parameter %v missing", requiredParams)
	}
	requiredParams = re.FindAllStringSubmatch(dR.WorkingDir, -1)
	if !isAllInParams(requiredParams, params) {
		return fmt.Errorf("Parameter %v missing", requiredParams)
	}
	for _, c := range dR.Command {
		requiredParams = re.FindAllStringSubmatch(c, -1)
		if !isAllInParams(requiredParams, params) {
			return fmt.Errorf("Parameter %v missing", requiredParams)
		}
	}

	for key, val := range dR.Environment {
		requiredParams = re.FindAllStringSubmatch(key, -1)
		if !isAllInParams(requiredParams, params) {
			return fmt.Errorf("Parameter %v missing", requiredParams)
		}
		requiredParams = re.FindAllStringSubmatch(val, -1)
		if !isAllInParams(requiredParams, params) {
			return fmt.Errorf("Parameter %v missing", requiredParams)
		}
	}
	return nil
}

func (dR *DockerRun) SetParams(params []domain.Parameter) error {
	dR.Parameters = params
	for _, param := range dR.Parameters {
		dR.Image = strings.Replace(dR.Image, fmt.Sprintf("${%s}", param.Name), param.Value, -1)
		dR.WorkingDir = strings.Replace(dR.WorkingDir, fmt.Sprintf("${%s}", param.Name), param.Value, -1)

		cmd := []string{}
		for _, c := range dR.Command {
			correctedCmd := strings.Replace(c, fmt.Sprintf("${%s}", param.Name), param.Value, -1)
			cmd = append(cmd, correctedCmd)
		}
		dR.Command = cmd

		env := map[string]string{}
		for key, val := range dR.Environment {
			correctedKey := strings.Replace(key, fmt.Sprintf("${%s}", param.Name), param.Value, -1)
			correctedVal := strings.Replace(val, fmt.Sprintf("${%s}", param.Name), param.Value, -1)
			env[correctedKey] = correctedVal
		}

		dR.Environment = env
	}
	return nil
}

func isAllInParams(matches [][]string, params []domain.Parameter) bool {
	for _, match := range matches {
		found := false
		for _, param := range params {
			if param.Name == match[1] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
