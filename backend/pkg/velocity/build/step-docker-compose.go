package build

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/docker"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"

	"github.com/ghodss/yaml"
	v3 "github.com/velocity-ci/velocity/backend/pkg/velocity/docker/compose/v3"

	"github.com/docker/docker/api/types/container"
)

type StepDockerCompose struct {
	BaseStep
	ComposeFilePath string `json:"composeFile"`
	// Contents    v3.DockerComposeYaml `json:"contents"`

	containerManager *docker.ContainerManager
}

func NewStepDockerCompose(c *v1.Step_DockerCompose, projectRoot string) *StepDockerCompose {
	streams, _ := getComposeFileStreams(
		filepath.Join(projectRoot, c.DockerCompose.GetComposeFile()),
	)

	return &StepDockerCompose{
		BaseStep:        newBaseStep("compose", streams),
		ComposeFilePath: c.DockerCompose.GetComposeFile(),
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
	contents, err := parseComposeFile(filepath.Join(t.ProjectRoot, dC.ComposeFilePath))
	if err != nil {
		return err
	}

	serviceOrder := v3.GetServiceOrder(contents.Services, []string{})

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

	dC.containerManager = docker.NewContainerManager(
		dC.ID,
		GetAuthConfigsMap(t.Docker.Registries),
		GetAddressAuthTokensMap(t.Docker.Registries),
	)

	for _, serviceName := range serviceOrder {
		writer := writers[serviceName]
		writer.SetStatus(StateBuilding)
		s := contents.Services[serviceName]

		// generate containerConfig + hostConfig
		containerConfig, hostConfig := dC.generateContainerAndHostConfig(
			s,
			serviceName, t.ProjectRoot)

		dC.containerManager.AddContainer(docker.NewContainer(
			writer,
			fmt.Sprintf("%s-%s", dC.ID, serviceName),
			s.Image,
			&s.Build,
			containerConfig,
			hostConfig,
			getServiceAliases(s.Networks["default"].Aliases, serviceName),
		))
	}

	if err := dC.containerManager.Execute(getSecrets(t.parameters)); err != nil {
		return err
	}

	if !dC.containerManager.IsSuccessful() {
		return fmt.Errorf("non-zero exit code")
	}

	return nil
}

func (dC *StepDockerCompose) Stop() error {
	if dC.containerManager != nil {
		dC.containerManager.Stop()
	}
	return nil
}

func (dC *StepDockerCompose) generateContainerAndHostConfig(
	s v3.DockerComposeService,
	serviceName,
	projectRoot string,
) (*container.Config, *container.HostConfig) {
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

	return containerConfig, hostConfig
}

func getServiceAliases(aliases []string, serviceName string) []string {
	for _, a := range aliases {
		if a == serviceName {
			return aliases
		}
	}

	return append(aliases, serviceName)
}
