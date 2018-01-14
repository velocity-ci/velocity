package velocity

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

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

func NewDockerCompose(y string) *DockerCompose {
	step := DockerCompose{
		BaseStep: BaseStep{
			Type: "compose",
		},
	}
	err := yaml.Unmarshal([]byte(y), &step)
	if err != nil {
		panic(err)
	}

	dir, _ := os.Getwd()
	dockerComposeYml, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s", dir, step.ComposeFile))
	err = yaml.Unmarshal(dockerComposeYml, &step.Contents)
	if err != nil {
		panic(err)
	}

	services := make([]string, len(step.Contents.Services))
	i := 0
	for k := range step.Contents.Services {
		services[i] = k
		i++
	}
	step.OutputStreams = services

	return &step
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

func (dC *DockerCompose) Execute(emitter Emitter, params map[string]Parameter) error {
	serviceOrder := getServiceOrder(dC.Contents.Services, []string{})

	services := []*serviceRunner{}
	var wg sync.WaitGroup
	cli, _ := client.NewEnvClient()
	ctx := context.Background()

	networkResp, err := cli.NetworkCreate(ctx, fmt.Sprintf("vci-%s", dC.GetRunID()), types.NetworkCreate{
		Labels: map[string]string{"owner": "velocity-ci"},
	})
	if err != nil {
		log.Println(err)
	}
	log.Println(networkResp.ID)

	writers := map[string]StreamWriter{}
	// Create writers
	for _, serviceName := range serviceOrder {
		writers[serviceName] = emitter.NewStreamWriter(serviceName)
	}

	for _, serviceName := range serviceOrder {
		writer := writers[serviceName]
		writer.SetStatus(StateRunning)
		writer.Write([]byte(fmt.Sprintf("Configured %s", serviceName)))

		s := dC.Contents.Services[serviceName]

		// generate containerConfig + hostConfig
		containerConfig, hostConfig := generateContainerAndHostConfig(s)

		// Create service runners
		sR := newServiceRunner(
			cli,
			ctx,
			writer,
			&wg,
			params,
			fmt.Sprintf("%s-%s", dC.GetRunID(), serviceName),
			s.Image,
			&s.Build,
			containerConfig,
			hostConfig,
			networkResp.ID,
		)

		services = append(services, sR)
	}

	// Pull/Build images
	for _, serviceRunner := range services {
		serviceRunner.PullOrBuild()
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
		log.Printf("network %s remove err: %s", networkResp.ID, err)
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
			writers[serviceName].Write([]byte(fmt.Sprintf("%s\n### FAILED \x1b[0m", errorANSI)))
		}
	} else {
		for _, serviceName := range serviceOrder {
			writers[serviceName].SetStatus(StateSuccess)
			writers[serviceName].Write([]byte(fmt.Sprintf("%s\n### SUCCESS \x1b[0m", successANSI)))
		}
	}

	return nil
}

func (dC *DockerCompose) String() string {
	j, _ := json.Marshal(dC)
	return string(j)
}

func generateContainerAndHostConfig(s dockerComposeService) (*container.Config, *container.HostConfig) {
	containerConfig := &container.Config{}
	if len(s.Command) > 0 {
		// containerConfig.Cmd = s.Command
	}
	return containerConfig, &container.HostConfig{}
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
	Image       string                    `json:"image" yaml:"image"`
	Build       dockerComposeServiceBuild `json:"build" yaml:"build"`
	WorkingDir  string                    `json:"workingDir" yaml:"working_dir"`
	Command     []string                  `json:"command" yaml:"command"`
	Links       []string                  `json:"links" yaml:"links"`
	Environment map[string]string         `json:"environment" yaml:"environment"`
	Volumes     []string                  `json:"volumes" yaml:"volumes"`
	Expose      []string                  `json:"expose" yaml:"expose"`
}

func (a *dockerComposeService) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var serviceMap map[string]interface{}
	err := unmarshal(&serviceMap)
	if err != nil {
		log.Printf("unable to unmarshal service")
		return err
	}
	// log.Printf("serviceMap: %+v\n", serviceMap)

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
		// a.Image = x.(string)
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
		log.Println("no environment specified")
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

	return nil
}

type dockerComposeServiceBuild struct {
	Context    string `json:"context" yaml:"context"`
	Dockerfile string `json:"dockerfile" yaml:"dockerfile"`
}
