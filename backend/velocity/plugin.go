package velocity

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Plugin struct {
	BaseStep
	Image          string            `json:"image" yaml:"image"`
	DockerInDocker bool              `json:"dind" yaml:"dind"`
	Environment    map[string]string `json:"environment" yaml:"environment"`
}

func (p Plugin) GetDetails() string {
	return fmt.Sprintf("image: %s dind: %v", p.Image, p.DockerInDocker)
}

func (p *Plugin) Execute(emitter Emitter, params map[string]Parameter) error {
	writer := emitter.NewStreamWriter("plugin")
	writer.SetStatus(StateRunning)
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

	var wg sync.WaitGroup
	cli, _ := client.NewEnvClient()
	ctx := context.Background()

	networkResp, err := cli.NetworkCreate(ctx, "", types.NetworkCreate{
		Labels: map[string]string{"owner": "velocity-ci"},
	})
	if err != nil {
		log.Println(err)
	}

	sR := newServiceRunner(
		cli,
		ctx,
		writer,
		&wg,
		params,
		fmt.Sprintf("%s-%s", p.GetRunID(), "plugin"),
		p.Image,
		nil,
		config,
		hostConfig,
		nil,
		networkResp.ID,
	)

	sR.PullOrBuild()
	sR.Create()
	stopServicesChannel := make(chan string, 32)
	wg.Add(1)
	go sR.Run(stopServicesChannel)
	_ = <-stopServicesChannel
	sR.Stop()

	wg.Wait()
	err = cli.NetworkRemove(ctx, networkResp.ID)
	if err != nil {
		log.Printf("network %s remove err: %s", networkResp.ID, err)
	}

	exitCode := sR.exitCode

	if err != nil {
		return err
	}

	if exitCode != 0 {
		writer.SetStatus("failed")
		writer.Write([]byte(fmt.Sprintf("%s\n### FAILED (exited: %d)\x1b[0m", errorANSI, exitCode)))
		return fmt.Errorf("Non-zero exit code: %d", exitCode)
	}

	writer.SetStatus("success")
	writer.Write([]byte(fmt.Sprintf("%s\n### SUCCESS (exited: %d)\x1b[0m", successANSI, exitCode)))
	return nil
}

func (p Plugin) Validate(params map[string]Parameter) error {

	return nil
}

func (p *Plugin) SetParams(params map[string]Parameter) error {

	return nil
}
