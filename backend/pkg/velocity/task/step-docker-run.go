package task

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	v3 "github.com/velocity-ci/velocity/backend/pkg/velocity/docker/compose/v3"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
)

type DockerRun struct {
	BaseStep
	Image          string                             `json:"image"`
	Command        v3.DockerComposeServiceCommand     `json:"command"`
	Environment    v3.DockerComposeServiceEnvironment `json:"environment"`
	WorkingDir     string                             `json:"workingDir"`
	MountPoint     string                             `json:"mountPoint"`
	IgnoreExitCode bool                               `json:"ignoreExitCode"`
}

func NewDockerRun() *DockerRun {
	return &DockerRun{
		Image:          "",
		Command:        []string{},
		Environment:    map[string]string{},
		WorkingDir:     "",
		MountPoint:     "",
		IgnoreExitCode: false,
		BaseStep:       newBaseStep("run", []string{"run"}),
	}
}

func (dR DockerRun) GetDetails() string {
	return fmt.Sprintf("image: %s command: %s", dR.Image, dR.Command)
}

func (dR *DockerRun) Execute(emitter out.Emitter, t *Task) error {
	writer := emitter.GetStreamWriter("run")
	defer writer.Close()
	writer.SetStatus(StateRunning)
	fmt.Fprintf(writer, out.ColorFmt(out.ANSIInfo, "-> %s"), dR.Description)

	if dR.MountPoint == "" {
		dR.MountPoint = "/velocity_ci"
	}
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
			fmt.Sprintf("%s:%s", dR.ProjectRoot, dR.MountPoint),
		},
	}

	var wg sync.WaitGroup
	cli, _ := client.NewEnvClient()
	ctx := context.Background()

	networkResp, err := cli.NetworkCreate(ctx, fmt.Sprintf("vci-%s", dR.GetRunID()), types.NetworkCreate{
		Labels: map[string]string{"owner": "velocity-ci"},
	})
	if err != nil {
		logging.GetLogger().Error("could not create docker network", zap.Error(err))
	}

	sR := newServiceRunner(
		cli,
		ctx,
		writer,
		&wg,
		t.ResolvedParameters,
		fmt.Sprintf("%s-%s", dR.GetRunID(), "run"),
		dR.Image,
		nil,
		config,
		hostConfig,
		nil,
		networkResp.ID,
	)

	sR.PullOrBuild(t.Docker.Registries)
	sR.Create()
	stopServicesChannel := make(chan string, 32)
	wg.Add(1)
	go sR.Run(stopServicesChannel)
	_ = <-stopServicesChannel
	sR.Stop()
	wg.Wait()
	err = cli.NetworkRemove(ctx, networkResp.ID)
	if err != nil {
		GetLogger().Error("could not remove docker network", zap.String("networkID", networkResp.ID), zap.Error(err))
	}

	exitCode := sR.exitCode

	if err != nil {
		return err
	}

	if exitCode != 0 && !dR.IgnoreExitCode {
		writer.SetStatus(StateFailed)
		fmt.Fprintf(writer, colorFmt(ansiError, "-> error (exited: %d)"), exitCode)

		return fmt.Errorf("Non-zero exit code: %d", exitCode)
	}

	writer.SetStatus(StateSuccess)
	fmt.Fprintf(writer, colorFmt(ansiSuccess, "-> success (exited: %d)"), exitCode)

	return nil
}

func (dR DockerRun) Validate(params map[string]Parameter) error {
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

func (dR *DockerRun) SetParams(params map[string]Parameter) error {
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
