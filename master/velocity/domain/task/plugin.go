package task

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
)

type Plugin struct {
	BaseStep
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

	exitCode, err := runContainer(
		resolvePullImage(p.Image),
		config,
		hostConfig,
		p.Parameters,
		p.Emit,
	)

	if err != nil {
		return err
	}

	if exitCode != 0 {
		p.Emit(fmt.Sprintf("%s\n### FAILED (exited: %d)\x1b[0m", errorANSI, exitCode))
		return fmt.Errorf("Non-zero exit code: %d", exitCode)
	}

	p.Emit(fmt.Sprintf("%s\n### SUCCESS (exited: %d)\x1b[0m", successANSI, exitCode))
	return nil
}

func (p Plugin) Validate(params []Parameter) error {

	return nil
}

func (p *Plugin) SetParams(params []Parameter) error {

	return nil
}
