package build

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/docker"

	"github.com/docker/docker/api/types/network"
	"github.com/ghodss/yaml"
	v3 "github.com/velocity-ci/velocity/backend/pkg/velocity/docker/compose/v3"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"go.uber.org/zap"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type StepDockerCompose struct {
	BaseStep
	ComposeFilePath string `json:"composeFile"`
	// Contents    v3.DockerComposeYaml `json:"contents"`
}

func NewStepDockerCompose(c *config.StepDockerCompose, projectRoot string) *StepDockerCompose {
	streams, _ := getComposeFileStreams(filepath.Join(projectRoot, c.ComposeFile))

	return &StepDockerCompose{
		BaseStep:        newBaseStep("compose", streams),
		ComposeFilePath: c.ComposeFile,
	}
}

func (dC StepDockerCompose) GetDetails() string {
	type details struct {
		ComposeFilePath string `json:"composeFile"`
	}
	y, _ := yaml.Marshal(&details{
		ComposeFilePath: dC.ComposeFilePath,
	})
	return string(y)
}

func (dC *StepDockerCompose) Validate(params map[string]Parameter) error {
	return nil
}

func (dC *StepDockerCompose) SetParams(params map[string]*Parameter) error {
	return nil
}

func parseComposeFile(path string) (*v3.DockerComposeYaml, error) {
	dockerComposeYml, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var contents v3.DockerComposeYaml
	err = yaml.Unmarshal(dockerComposeYml, &contents)
	if err != nil {
		return nil, err
	}

	return &contents, nil
}

func getComposeFileStreams(path string) ([]string, error) {
	contents, err := parseComposeFile(path)
	if err != nil {
		return nil, err
	}
	services := make([]string, len(contents.Services))
	i := 0
	for k := range contents.Services {
		services[i] = k
		i++
	}

	return services, nil
}

func (dC *StepDockerCompose) Execute(emitter Emitter, t *Task) error {

	// err := dC.parseDockerComposeFile(t.ProjectRoot)
	contents, err := parseComposeFile(filepath.Join(t.ProjectRoot, dC.ComposeFilePath))
	if err != nil {
		return err
	}

	serviceOrder := v3.GetServiceOrder(contents.Services, []string{})

	services := []*docker.ServiceRunner{}
	var wg sync.WaitGroup
	cli, _ := client.NewEnvClient()
	ctx := context.Background()

	networkResp, err := cli.NetworkCreate(ctx, fmt.Sprintf("vci-%s", dC.ID), types.NetworkCreate{
		Labels: map[string]string{"owner": "velocity-ci"},
	})
	if err != nil {
		logging.GetLogger().Error("could not create docker network", zap.String("err", err.Error()))
	}

	writers := map[string]StreamWriter{}
	// Create writers
	for _, serviceName := range serviceOrder {
		serviceWriter, err := dC.GetStreamWriter(emitter, serviceName)
		if err != nil {
			return err
		}
		writers[serviceName] = serviceWriter
		defer writers[serviceName].Close()
	}

	for _, serviceName := range serviceOrder {
		writer := writers[serviceName]
		writer.SetStatus(StateBuilding)
		s := contents.Services[serviceName]

		// generate containerConfig + hostConfig
		containerConfig, hostConfig, networkConfig := dC.generateContainerAndHostConfig(s, serviceName, networkResp.ID, t.ProjectRoot)

		// Create service runners
		sR := docker.NewServiceRunner(
			cli,
			ctx,
			writer,
			&wg,
			getSecrets(t.parameters),
			fmt.Sprintf("%s-%s", dC.ID, serviceName),
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
		serviceRunner.PullOrBuild(GetAuthConfigsMap(t.Docker.Registries), GetAddressAuthTokensMap(t.Docker.Registries))
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
		if serviceRunner.ExitCode != 0 {
			success = false

			break
		}
	}

	if !success {
		for _, serviceName := range serviceOrder {
			writers[serviceName].SetStatus(StateFailed)
			fmt.Fprintf(writers[serviceName], output.ColorFmt(output.ANSIError, "-> %s error", "\n"), serviceName)

		}
	} else {
		for _, serviceName := range serviceOrder {
			writers[serviceName].SetStatus(StateSuccess)
			fmt.Fprintf(writers[serviceName], output.ColorFmt(output.ANSISuccess, "-> %s success", "\n"), serviceName)

		}
	}

	return nil
}

func (dC *StepDockerCompose) GracefulStop() error {
	return nil
}

func (dC *StepDockerCompose) generateContainerAndHostConfig(
	s v3.DockerComposeService,
	serviceName,
	networkID,
	projectRoot string,
) (*container.Config, *container.HostConfig, *network.NetworkingConfig) {
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
				hostMount = filepath.Join(projectRoot, filepath.Dir(dC.ComposeFilePath), hostMount)
				if strings.Contains(hostMount, projectRoot) { // no further up from project root
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
			target = docker.GetContainerName(fmt.Sprintf("%s-%s", dC.ID, l))
			alias = l
		} else {
			target = parts[0]
			target = docker.GetContainerName(fmt.Sprintf("%s-%s", dC.ID, target))
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
