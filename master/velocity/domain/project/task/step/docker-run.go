package step

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/velocity-ci/velocity/master/velocity/domain"
)

type DockerRun struct {
	BaseStep
	Description string            `json:"description" yaml:"description"`
	Image       string            `json:"image" yaml:"image"`
	Command     []string          `json:"command" yaml:"command"`
	Environment map[string]string `json:"environment" yaml:"environment"`
	WorkingDir  string            `json:"workingDir" yaml:"working_dir"`
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

func (dB DockerRun) Execute() error {
	cli, err := client.NewEnvClient()
	ctx := context.Background()
	if err != nil {
		panic(err)
	}
	// Pull Image
	reader, err := cli.ImagePull(ctx, fmt.Sprintf("docker.io/%s", dB.Image), types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	// Create and run container
	env := []string{}
	for k, v := range dB.Environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cwd, _ := os.Getwd()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: dB.Image,
		Cmd:   dB.Command,
		Volumes: map[string]struct{}{
			"/velocity_ci": struct{}{},
		},
		WorkingDir: fmt.Sprintf("/velocity_ci/%s", dB.WorkingDir),
		Env:        env,
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/velocity_ci", cwd),
		},
	}, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		panic(err)
	}
	scanner = bufio.NewScanner(out)
	for scanner.Scan() {
		allBytes := scanner.Bytes()
		fmt.Println(string(allBytes[8:]))
	}

	if _, err = cli.ContainerWait(ctx, resp.ID); err != nil {
		panic(err)
	}

	cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{})

	io.Copy(os.Stdout, out)
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
	for _, param := range params {
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
