package velocity

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/docker/docker/client"
	yaml "gopkg.in/yaml.v2"
)

type DockerCompose struct {
	BaseStep
	ComposeFile string `json:"composeFile" yaml:"compose_file"`
	Contents    dockerComposeYaml
}

func NewDockerCompose(y string) *DockerCompose {
	step := DockerCompose{
		BaseStep: BaseStep{
			Type: "compose",
		},
	}
	err := yaml.Unmarshal([]byte(y), step)
	if err != nil {
		panic(err)
	}

	dir, _ := os.Getwd()
	dockerComposeYml, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s", dir, step.ComposeFile))
	err = yaml.Unmarshal(dockerComposeYml, step.Contents)
	if err != nil {
		panic(err)
	}

	services := make([]string, 0, len(step.Contents.Services))
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

	for _, serviceName := range serviceOrder {
		writer := emitter.NewStreamWriter(serviceName)
		writer.SetStatus(StateRunning)
		writer.Write([]byte(fmt.Sprintf("Starting %s", serviceName)))

		services = append(services)

		go startService(&services, serviceName)
	}
	return nil
}

func NewServiceRunner() *serviceRunner {
	return &serviceRunner{}
}

type serviceRunner struct {
	writer      io.Writer
	containerID string
	exitCode    int
	wg          *sync.WaitGroup
}

func (sR *serviceRunner) Start() {
	cli, err := client.NewEnvClient()
	if err != nil {

	}
}

func (sR *serviceRunner) Stop() {

}

func stopAll(services []*serviceRunner) {
	for _, s := range services {
		s.Stop()
	}
}

func startService(services *[]*serviceRunner, serviceName string) {

}

func (dC *DockerCompose) String() string {
	j, _ := json.Marshal(dC)
	return string(j)
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
	Build       dockerComposeServiceBuild `json:"build" yaml:"build"`
	WorkingDir  string                    `json:"workingDir" yaml:"working_dir"`
	Command     string                    `json:"command" yaml:"command"`
	Links       []string                  `json:"links" yaml:"links"`
	Environment map[string]string         `json:"environment" yaml:"environment"`
	Volumes     []string                  `json:"volumes" yaml:"volumes"`
	Expose      []string                  `json:"expose" yaml:"expose"`
}

type dockerComposeServiceBuild struct {
	Context    string `json:"context" yaml:"context"`
	Dockerfile string `json:"dockerfile" yaml:"dockerfile"`
}
