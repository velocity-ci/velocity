package step

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/velocity-ci/velocity/master/velocity/domain"
)

const successANSI = "\x1b[1m\x1b[40m\x1b[32m"
const errorANSI = "\x1b[1m\x1b[40m\x1b[31m"
const infoANSI = "\x1b[1m\x1b[40m\x1b[34m"

type DockerRun struct {
	domain.BaseStep `yaml:",inline"`
	Image           string            `json:"image" yaml:"image"`
	Command         []string          `json:"command" yaml:"command"`
	Environment     map[string]string `json:"environment" yaml:"environment"`
	WorkingDir      string            `json:"workingDir" yaml:"working_dir"`
	MountPoint      string            `json:"mountPoint" yaml:"mount_point"`
	IgnoreExitCode  bool              `json:"ignoreExitCode" yaml:"ignore_exit"`
}

func (dR *DockerRun) SetEmitter(e func(string)) {
	dR.Emit = e
}

func (dR DockerRun) GetType() string {
	return "run"
}

func (dR DockerRun) GetDescription() string {
	return dR.Description
}

func (dR DockerRun) GetDetails() string {
	return fmt.Sprintf("image: %s command: %s", dR.Image, dR.Command)
}

func (dR *DockerRun) Execute() error {

	dR.Emit(fmt.Sprintf("%s\n## %s\n\x1b[0m", infoANSI, dR.Description))

	if dR.MountPoint == "" {
		dR.MountPoint = "/velocity_ci"
	}
	env := []string{}
	for k, v := range dR.Environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cwd, _ := os.Getwd()
	if os.Getenv("SIB_CWD") != "" {
		cwd = os.Getenv("SIB_CWD")
	}

	config := &container.Config{
		Image: dR.Image,
		Cmd:   dR.Command,
		Volumes: map[string]struct{}{
			dR.MountPoint: struct{}{},
		},
		WorkingDir: fmt.Sprintf("%s/%s", dR.MountPoint, dR.WorkingDir),
		Env:        env,
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/%s", cwd, dR.MountPoint),
		},
	}

	exitCode, err := runContainer(
		resolvePullImage(dR.Image),
		config,
		hostConfig,
		dR.Parameters,
		dR.Emit,
	)

	if err != nil {
		return err
	}

	if exitCode != 0 && !dR.IgnoreExitCode {
		dR.Emit(fmt.Sprintf("%s\n### FAILED (exited: %d)\x1b[0m", errorANSI, exitCode))
		return fmt.Errorf("Non-zero exit code: %d", exitCode)
	}

	dR.Emit(fmt.Sprintf("%s\n### SUCCESS (exited: %d)\x1b[0m", successANSI, exitCode))
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