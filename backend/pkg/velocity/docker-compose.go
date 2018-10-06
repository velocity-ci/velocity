package velocity

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/network"
	"go.uber.org/zap"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	yaml "gopkg.in/yaml.v2"
)

type DockerCompose struct {
	BaseStep
	ComposeFile string `json:"composeFile" yaml:"composeFile"`
	Contents    dockerComposeYaml
}

func NewDockerCompose() *DockerCompose {
	return &DockerCompose{
		BaseStep: newBaseStep("compose", []string{}),
	}
}

func (s *DockerCompose) UnmarshalYamlInterface(y map[interface{}]interface{}) error {
	switch x := y["composeFile"].(type) {
	case interface{}:
		s.ComposeFile = x.(string)
		break
	}
	err := s.BaseStep.UnmarshalYamlInterface(y)
	if err != nil {
		return err
	}
	return s.parseDockerComposeFile()
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

func (dC *DockerCompose) Execute(emitter Emitter, t *Task) error {

	err := dC.parseDockerComposeFile()
	if err != nil {
		return err
	}

	serviceOrder := getServiceOrder(dC.Contents.Services, []string{})

	services := []*serviceRunner{}
	var wg sync.WaitGroup
	cli, _ := client.NewEnvClient()
	ctx := context.Background()

	networkResp, err := cli.NetworkCreate(ctx, fmt.Sprintf("vci-%s", dC.GetRunID()), types.NetworkCreate{
		Labels: map[string]string{"owner": "velocity-ci"},
	})
	if err != nil {
		GetLogger().Error("could not create docker network", zap.String("err", err.Error()))
	}

	writers := map[string]StreamWriter{}
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
		containerConfig, hostConfig, networkConfig := dC.generateContainerAndHostConfig(s, networkResp.ID)

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
		GetLogger().Error("could not remove docker network", zap.String("networkID", networkResp.ID), zap.Error(err))
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
			fmt.Fprintf(writers[serviceName], colorFmt(ansiError, "-> %s error"), serviceName)

		}
	} else {
		for _, serviceName := range serviceOrder {
			writers[serviceName].SetStatus(StateSuccess)
			fmt.Fprintf(writers[serviceName], colorFmt(ansiSuccess, "-> %s success"), serviceName)

		}
	}

	return nil
}

func (dC *DockerCompose) String() string {
	j, _ := json.Marshal(dC)
	return string(j)
}

func (dC *DockerCompose) generateContainerAndHostConfig(s dockerComposeService, networkID string) (*container.Config, *container.HostConfig, *network.NetworkingConfig) {
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
		Cmd:        s.Command,
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
				Aliases: s.Networks["default"].Aliases,
			},
		},
	}

	return containerConfig, hostConfig, networkConfig
}

func getServiceOrder(services map[string]dockerComposeService, serviceOrder []string) []string {
	for serviceName, serviceDef := range services {
		if isIn(serviceName, serviceOrder) {
			break
		}
		for _, linkedService := range serviceDef.Links {
			serviceOrder = getLinkedServiceOrder(linkedService, services, serviceOrder)
		}
		serviceOrder = append(serviceOrder, serviceName)
	}

	for len(services) != len(serviceOrder) {
		serviceOrder = getServiceOrder(services, serviceOrder)
	}

	return serviceOrder
}

func getLinkedServiceOrder(serviceName string, services map[string]dockerComposeService, serviceOrder []string) []string {
	if isIn(serviceName, serviceOrder) {
		return serviceOrder
	}
	for _, linkedService := range services[serviceName].Links {
		serviceOrder = getLinkedServiceOrder(linkedService, services, serviceOrder)
	}
	return append(serviceOrder, serviceName)
}

func isIn(needle string, haystack []string) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}
	return false
}

type dockerComposeYaml struct {
	Services map[string]dockerComposeService `json:"services" yaml:"services"`
}

type dockerComposeService struct {
	Image       string                                 `json:"image" yaml:"image"`
	Build       dockerComposeServiceBuild              `json:"build" yaml:"build"`
	WorkingDir  string                                 `json:"workingDir" yaml:"working_dir"`
	Command     []string                               `json:"command" yaml:"command"`
	Links       []string                               `json:"links" yaml:"links"`
	Environment map[string]string                      `json:"environment" yaml:"environment"`
	Volumes     []string                               `json:"volumes" yaml:"volumes"`
	Expose      []string                               `json:"expose" yaml:"expose"`
	Networks    map[string]dockerComposeServiceNetwork `json:"networks" yaml:"networks"`
}

type dockerComposeServiceNetwork struct {
	Aliases []string `json:"aliases" yaml:"aliases"`
}

type dockerComposeServiceBuild struct {
	Context    string `json:"context" yaml:"context"`
	Dockerfile string `json:"dockerfile" yaml:"dockerfile"`
}

func (a *dockerComposeService) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var serviceMap map[string]interface{}
	err := unmarshal(&serviceMap)
	if err != nil {
		GetLogger().Error("could not unmarshal service", zap.Error(err))
		return err
	}

	// image
	switch x := serviceMap["image"].(type) {
	case interface{}:
		a.Image = x.(string)
		break
	default:
		break
	}

	// build
	switch x := serviceMap["build"].(type) {
	case string:
		// use string as context path. Dockerfile in root of that path
		a.Build = dockerComposeServiceBuild{
			Context:    x, // get path of docker-compose file
			Dockerfile: "Dockerfile",
		}
		break
	case map[interface{}]interface{}:
		a.Build = dockerComposeServiceBuild{
			Context:    x["context"].(string), // get path of docker-compose file
			Dockerfile: x["dockerfile"].(string),
		}
		break
	default:
		break
	}

	// command
	switch x := serviceMap["command"].(type) {
	case []interface{}:
		for _, p := range x {
			a.Command = append(a.Command, p.(string))
		}
		break
	case interface{}:
		// TODO: handle /bin/sh -c "sleep 3"; should be: ["/bin/sh", "-c", "\"sleep 3\""]
		a.Command = strings.Split(x.(string), " ")
		break
	default:
		break
	}

	// working_dir
	switch x := serviceMap["working_dir"].(type) {
	case interface{}:
		a.WorkingDir = x.(string)
		break
	default:
		break
	}

	// environment
	a.Environment = map[string]string{}
	switch x := serviceMap["environment"].(type) {
	case []interface{}:
		for _, e := range x {
			parts := strings.Split(e.(string), "=")
			key := parts[0]
			val := parts[1]
			a.Environment[key] = val
		}
		break
	case map[interface{}]interface{}:
		for k, v := range x {
			if num, ok := v.(int); ok {
				v = strconv.Itoa(num)
			}
			a.Environment[k.(string)] = v.(string)
		}
		break
	default:
		break
	}

	// volumes
	switch x := serviceMap["volumes"].(type) {
	case []interface{}:
		for _, v := range x {
			a.Volumes = append(a.Volumes, v.(string))
		}
		break
	}

	// links
	switch x := serviceMap["links"].(type) {
	case []interface{}:
		for _, v := range x {
			a.Links = append(a.Links, v.(string))
		}
		break
	}

	// expose
	switch x := serviceMap["expose"].(type) {
	case []interface{}:
		for _, v := range x {
			a.Expose = append(a.Expose, v.(string))
		}
		break
	}

	// networks
	switch x := serviceMap["networks"].(type) {
	case map[interface{}]interface{}:
		d := x["default"].(map[interface{}]interface{})
		iA := d["aliases"].([]interface{})
		aliases := []string{}
		for _, i := range iA {
			aliases = append(aliases, i.(string))
		}
		a.Networks = map[string]dockerComposeServiceNetwork{
			"default": {
				Aliases: aliases,
			},
		}
		break
	default:
		break
	}

	return nil
}
