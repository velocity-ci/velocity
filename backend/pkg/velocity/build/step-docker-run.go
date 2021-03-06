package build

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/ghodss/yaml"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/docker"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
)

type StepDockerRun struct {
	BaseStep
	Image          string            `json:"image"`
	Command        []string          `json:"command"`
	Environment    map[string]string `json:"environment"`
	WorkingDir     string            `json:"workingDir"`
	MountPoint     string            `json:"mountPoint"`
	IgnoreExitCode bool              `json:"ignoreExitCode"`

	containerManager *docker.ContainerManager
}

func NewStepDockerRun(c *config.StepDockerRun) *StepDockerRun {
	if c.MountPoint == "" {
		c.MountPoint = "/velocity_ci"
	}
	if c.Environment == nil {
		c.Environment = map[string]string{}
	}
	return &StepDockerRun{
		BaseStep:       newBaseStep("run", []string{"run"}),
		Image:          c.Image,
		Command:        c.Command,
		Environment:    c.Environment,
		WorkingDir:     c.WorkingDir,
		MountPoint:     c.MountPoint,
		IgnoreExitCode: c.IgnoreExitCode,
	}
}

func (dR StepDockerRun) GetDetails() string {
	type details struct {
		Image          string            `json:"image"`
		Command        string            `json:"command"`
		Environment    map[string]string `json:"environment"`
		WorkingDir     string            `json:"workingDir"`
		MountPoint     string            `json:"mountPoint"`
		IgnoreExitCode bool              `json:"ignoreExitCode"`
	}
	y, _ := yaml.Marshal(&details{
		Image:          dR.Image,
		Command:        strings.Join(dR.Command, " "),
		Environment:    dR.Environment,
		WorkingDir:     dR.WorkingDir,
		MountPoint:     dR.MountPoint,
		IgnoreExitCode: dR.IgnoreExitCode,
	})
	return string(y)
}

func (dR *StepDockerRun) Execute(emitter Emitter, t *Task) error {
	writer, err := dR.GetStreamWriter(emitter, "run")
	if err != nil {
		return err
	}
	defer writer.Close()
	writer.SetStatus(StateBuilding)
	fmt.Fprintf(writer, "\r")

	env := []string{}
	for k, v := range dR.Environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	// Only used for Docker-based CLI. Unsupported right now.
	// if os.Getenv("SIB_CWD") != "" {
	// 	cwd = os.Getenv("SIB_CWD")
	// }

	config := &container.Config{
		Image: dR.Image,
		Cmd:   []string(dR.Command),
		Volumes: map[string]struct{}{
			dR.MountPoint: {},
		},
		WorkingDir: fmt.Sprintf("%s/%s", dR.MountPoint, dR.WorkingDir),
		Env:        env,
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:%s", t.ProjectRoot, dR.MountPoint),
		},
	}

	dR.containerManager = docker.NewContainerManager(
		dR.ID,
		GetAuthConfigsMap(t.Docker.Registries),
		GetAddressAuthTokensMap(t.Docker.Registries),
	)

	dR.containerManager.AddContainer(docker.NewContainer(
		writer,
		fmt.Sprintf("%s-%s", dR.ID, "run"),
		dR.Image,
		nil,
		config,
		hostConfig,
		nil,
	))

	if err := dR.containerManager.Execute(getSecrets(t.parameters)); err != nil {
		return err
	}

	if !dR.containerManager.IsSuccessful() && !dR.IgnoreExitCode {
		writer.SetStatus(StateFailed)
		fmt.Fprintf(writer, output.ColorFmt(output.ANSIError, "-> error: non-zero exit code", "\n"))

		return fmt.Errorf("non-zero exit code")
	}

	writer.SetStatus(StateSuccess)
	fmt.Fprintf(writer, output.ColorFmt(output.ANSISuccess, "-> success", "\n"))

	return nil
}

func (dR *StepDockerRun) Stop() error {
	if dR.containerManager != nil {
		dR.containerManager.Stop()
	}
	return nil
}

func (dR StepDockerRun) Validate(params map[string]Parameter) error {
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

func (dR *StepDockerRun) SetParams(params map[string]*Parameter) error {
	for paramName, param := range params {
		dR.Image = strings.Replace(dR.Image, fmt.Sprintf("${%s}", paramName), param.Value, -1)
		dR.WorkingDir = strings.Replace(dR.WorkingDir, fmt.Sprintf("${%s}", paramName), param.Value, -1)

		cmd := []string{}
		for _, c := range dR.Command {
			correctedCmd := strings.Replace(c, fmt.Sprintf("${%s}", paramName), param.Value, -1)
			cmd = append(cmd, correctedCmd)
		}
		dR.Command = cmd

		env := map[string]string{}
		for key, val := range dR.Environment {
			correctedKey := strings.Replace(key, fmt.Sprintf("${%s}", paramName), param.Value, -1)
			correctedVal := strings.Replace(val, fmt.Sprintf("${%s}", paramName), param.Value, -1)
			env[correctedKey] = correctedVal
		}

		dR.Environment = env
	}
	return nil
}

func isAllInParams(matches [][]string, params map[string]Parameter) bool {
	for _, match := range matches {
		found := false
		for paramName := range params {
			if paramName == match[1] {
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
