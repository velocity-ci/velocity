package task

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/network"
	"github.com/ghodss/yaml"
	v3 "github.com/velocity-ci/velocity/backend/pkg/velocity/docker/compose/v3"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
	"go.uber.org/zap"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerCompose struct {
	BaseStep
	ComposeFile string `json:"composeFile"`
	Contents    v3.DockerComposeYaml
}

func NewDockerCompose() *DockerCompose {
	return &DockerCompose{
		BaseStep: newBaseStep("compose", []string{}),
	}
}

func (dC DockerCompose) GetDetails() string {
	return fmt.Sprintf("composeFile: %s", dC.ComposeFile)
}

func (dC *DockerCompose) Validate(params map[string]Parameter) error {
	return nil
}

func (dC *DockerCompose) SetParams(params map[string]Parameter) error {
	return nil
}

func (dC *DockerCompose) parseDockerComposeFile() error {
	dockerComposeYml, err := ioutil.ReadFile(filepath.Join(dC.ProjectRoot, dC.ComposeFile))
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(dockerComposeYml, &dC.Contents)
	if err != nil {
		return err
	}

	services := make([]string, len(dC.Contents.Services))
	i := 0
	for k := range dC.Contents.Services {
		services[i] = k
		i++
	}
	dC.OutputStreams = services

	return nil
}

func (dC *DockerCompose) Execute(emitter out.Emitter, t *Task) error {

	err := dC.parseDockerComposeFile()
	if err != nil {
		return err
	}

	serviceOrder := v3.GetServiceOrder(dC.Contents.Services, []string{})

	services := []*serviceRunner{}
	var wg sync.WaitGroup
	cli, _ := client.NewEnvClient()
	ctx := context.Background()

	networkResp, err := cli.NetworkCreate(ctx, fmt.Sprintf("vci-%s", dC.GetRunID()), types.NetworkCreate{
		Labels: map[string]string{"owner": "velocity-ci"},
	})
	if err != nil {
		logging.GetLogger().Error("could not create docker network", zap.String("err", err.Error()))
	}

	writers := map[string]out.StreamWriter{}
	// Create writers
	for _, serviceName := range serviceOrder {
		writers[serviceName] = emitter.GetStreamWriter(serviceName)
		defer writers[serviceName].Close()
	}

	for _, serviceName := range serviceOrder {
		writer := writers[serviceName]
		writer.SetStatus(StateRunning)
		s := dC.Contents.Services[serviceName]

		// generate containerConfig + hostConfig
		containerConfig, hostConfig, networkConfig := dC.generateContainerAndHostConfig(s, serviceName, networkResp.ID)

		// Create service runners
		sR := newServiceRunner(
			cli,
			ctx,
			writer,
			&wg,
			t.ResolvedParameters,
			fmt.Sprintf("%s-%s", dC.GetRunID(), serviceName),
			s.Image,
			&s.Build,
			containerConfig,
			hostConfig,
			networkConfig,
			networkResp.ID,
		)

		services = append(services, sR)
	}

	// Pull/Build images
	for _, serviceRunner := range services {
		serviceRunner.PullOrBuild(t.Docker.Registries)
	}

	// Create services
	for _, serviceRunner := range services {
		serviceRunner.Create()
	}
	stopServicesChannel := make(chan string, 32)
	// Start services
	for _, serviceRunner := range services {
		wg.Add(1)
		go serviceRunner.Run(stopServicesChannel)
	}

	_ = <-stopServicesChannel
	for _, s := range services {
		s.Stop()
	}
	wg.Wait()
	err = cli.NetworkRemove(ctx, networkResp.ID)
	if err != nil {
		logging.GetLogger().Error("could not remove docker network", zap.String("networkID", networkResp.ID), zap.Error(err))
	}
	success := true
	for _, serviceRunner := range services {
		if serviceRunner.exitCode != 0 {
			success = false

			break
		}
	}

	if !success {
		for _, serviceName := range serviceOrder {
			writers[serviceName].SetStatus(StateFailed)
			fmt.Fprintf(writers[serviceName], out.ColorFmt(out.ANSIError, "-> %s error"), serviceName)

		}
	} else {
		for _, serviceName := range serviceOrder {
			writers[serviceName].SetStatus(StateSuccess)
			fmt.Fprintf(writers[serviceName], out.ColorFmt(out.ANSISuccess, "-> %s success"), serviceName)

		}
	}

	return nil
}

func (dC *DockerCompose) generateContainerAndHostConfig(s v3.DockerComposeService, serviceName, networkID string) (*container.Config, *container.HostConfig, *network.NetworkingConfig) {
	env := []string{}
	for k, v := range s.Environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	volumes := map[string]struct{}{}
	binds := []string{}
	for _, v := range s.Volumes {
		parts := strings.Split(v, ":")
		if len(parts) == 1 {
			volumes[parts[0]] = struct{}{}
		} else if len(parts) > 1 {
			hostMount := parts[0]
			guestMount := parts[1:]
			volumes[parts[1]] = struct{}{}
			if !filepath.IsAbs(hostMount) { // no absolute paths allowed.
				hostMount = filepath.Join(dC.ProjectRoot, filepath.Dir(dC.ComposeFile), hostMount)
				if strings.Contains(hostMount, dC.ProjectRoot) { // no further up from project root
					binds = append(binds, strings.Join(append([]string{hostMount}, guestMount...), ":"))
				}
			}
		}
	}

	containerConfig := &container.Config{
		Image:      s.Image,
		Cmd:        []string(s.Command),
		Env:        env,
		Volumes:    volumes,
		WorkingDir: s.WorkingDir,
	}

	links := []string{}
	for _, l := range s.Links {
		parts := strings.Split(l, ":")
		var target string
		var alias string
		if len(parts) == 1 {
			target = getContainerName(fmt.Sprintf("%s-%s", dC.GetRunID(), l))
			alias = l
		} else {
			target = parts[0]
			target = getContainerName(fmt.Sprintf("%s-%s", dC.GetRunID(), target))
			alias = parts[1]
		}
		links = append(links, fmt.Sprintf("%s:%s", target, alias))
	}

	hostConfig := &container.HostConfig{
		Binds: binds,
		Links: links,
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkID: {
				Aliases: getServiceAliases(s.Networks["default"].Aliases, serviceName),
			},
		},
	}

	return containerConfig, hostConfig, networkConfig
}

func getServiceAliases(aliases []string, serviceName string) []string {
	for _, a := range aliases {
		if a == serviceName {
			return aliases
		}
	}

	return append(aliases, serviceName)
}
