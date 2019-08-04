package docker

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v3 "github.com/velocity-ci/velocity/backend/pkg/velocity/docker/compose/v3"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
	"go.uber.org/zap"
	"golang.org/x/net/http/httpproxy"
)

const owner = "velocity-ci"
const ownerPrefix = "vci"

type ContainerManager struct {
	wg        sync.WaitGroup
	mutex     sync.Mutex
	networkID string

	id          string
	authConfigs map[string]types.AuthConfig
	authTokens  map[string]string

	containers      []*Container
	running         bool
	firstStoppedSvc string
}

func NewContainerManager(
	id string,
	registryAuthConfigs map[string]types.AuthConfig,
	registryAuthTokens map[string]string,
) *ContainerManager {
	return &ContainerManager{
		id:          id,
		containers:  []*Container{},
		authConfigs: registryAuthConfigs,
		authTokens:  registryAuthTokens,
		running:     false,
	}
}

func (cM *ContainerManager) AddContainer(container *Container) error {
	cM.containers = append(cM.containers, container)
	return nil
}

func (cM *ContainerManager) doContainers(f func(c *Container) error) error {
	if cM.IsRunning() {
		cM.mutex.Lock()
		defer cM.mutex.Unlock()
		for _, c := range cM.containers {
			if err := f(c); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cM *ContainerManager) Execute(secrets []string) error {
	defer cM.Stop()
	cM.running = true

	if cM.IsRunning() {
		cM.mutex.Lock()
		networkResp, err := dockerClient.NetworkCreate(
			context.Background(),
			fmt.Sprintf("%s-%s", ownerPrefix, cM.id),
			types.NetworkCreate{
				Labels: map[string]string{"owner": owner},
			})
		if err != nil {
			logging.GetLogger().Error("could not create docker network", zap.Error(err))
			return err
		}

		cM.networkID = networkResp.ID
		cM.mutex.Unlock()
	}

	err := cM.doContainers(func(c *Container) error {
		return c.Get(secrets, cM.authConfigs, cM.authTokens)
	})
	if err != nil {
		return err
	}

	err = cM.doContainers(func(c *Container) error {
		return c.ConfigureNetwork(cM.networkID)
	})
	if err != nil {
		return err
	}

	err = cM.doContainers(func(c *Container) error {
		return c.Create()
	})
	if err != nil {
		return err
	}

	if cM.IsRunning() {
		cM.mutex.Lock()
		firstStoppedSvcCh := make(chan string, 1)
		// Start services
		for _, container := range cM.containers {
			cM.wg.Add(1)
			go container.Run(&cM.wg, secrets, firstStoppedSvcCh)
		}
		cM.mutex.Unlock()
		cM.firstStoppedSvc = <-firstStoppedSvcCh
	}

	if !cM.IsRunning() {
		return fmt.Errorf("containers interrupted")
	}

	if err := cM.Stop(); err != nil {
		return err
	}

	return nil
}

func (cM *ContainerManager) IsRunning() bool {
	cM.mutex.Lock()
	defer cM.mutex.Unlock()
	return cM.running
}

func (cM *ContainerManager) IsSuccessful() bool {
	cM.mutex.Lock()
	defer cM.mutex.Unlock()
	for _, c := range cM.containers {
		if c.name == cM.firstStoppedSvc {
			return c.exitCode == 0
		}
	}
	return false
}

func (cM *ContainerManager) IsAllSuccessful() bool {
	cM.mutex.Lock()
	defer cM.mutex.Unlock()
	for _, c := range cM.containers {
		if c.exitCode != 0 {
			return false
		}
	}
	return true
}

func (cM *ContainerManager) Stop() error {
	if cM.IsRunning() {
		cM.mutex.Lock()
		defer cM.mutex.Unlock()
		cM.running = false
		for _, container := range cM.containers {
			if err := container.Stop(); err != nil {
				return err
			}
		}
		cM.wg.Wait()
		if err := dockerClient.NetworkRemove(context.Background(), cM.networkID); err != nil {
			logging.GetLogger().Error("could not remove docker network", zap.String("networkID", cM.networkID), zap.Error(err))
			return err
		}
	}
	return nil
}

type Container struct {
	writer io.Writer

	name  string
	image string
	build *v3.DockerComposeServiceBuild

	containerConfig *container.Config
	hostConfig      *container.HostConfig
	networkConfig   *network.NetworkingConfig
	networkAliases  []string

	containerID string
	networkID   string
	running     bool
	exitCode    int
	mutex       sync.Mutex
}

func NewContainer(
	writer io.Writer,
	name string,
	image string,
	build *v3.DockerComposeServiceBuild,
	containerConfig *container.Config,
	hostConfig *container.HostConfig,
	networkAliases []string,
) *Container {
	return &Container{
		writer:          writer,
		name:            name,
		image:           image,
		build:           build,
		containerConfig: containerConfig,
		hostConfig:      hostConfig,
		networkAliases:  networkAliases,
		running:         false,
	}
}

func (c *Container) ConfigureNetwork(networkID string) error {
	c.networkConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkID: {
				Aliases: c.networkAliases,
			},
		},
	}
	return nil
}

func (c *Container) Get(secrets []string, authConfigs map[string]types.AuthConfig, authTokens map[string]string) error {
	if c.build != nil && (c.build.Dockerfile != "" || c.build.Context != "") {
		return c.Build(secrets, authConfigs, authTokens)
	}
	return c.Pull(secrets, authConfigs, authTokens)
}

func (c *Container) Build(
	secrets []string,
	authConfigs map[string]types.AuthConfig,
	addressAuthToken map[string]string,
) error {
	builder := NewImageBuilder()
	err := builder.Build(
		c.writer,
		secrets,
		c.build.Context,
		c.build.Dockerfile,
		[]string{GetImageName(c.name)},
		authConfigs,
	)
	if err != nil {
		logging.GetLogger().Error("could not build image", zap.String("err", err.Error()))
		return err
	}
	c.image = GetImageName(c.name)
	c.containerConfig.Image = GetImageName(c.name)
	return nil
}

func (c *Container) Pull(
	secrets []string,
	authConfigs map[string]types.AuthConfig,
	addressAuthToken map[string]string,
) error {
	// check if image exists locally before pulling
	c.image = resolvePullImage(c.image)
	c.containerConfig.Image = resolvePullImage(c.image)
	authToken := getAuthToken(c.image, addressAuthToken)
	pullResp, err := dockerClient.ImagePull(
		context.Background(),
		c.image,
		types.ImagePullOptions{
			RegistryAuth: authToken,
		},
	)
	if err != nil {
		logging.GetLogger().Error("could not pull image", zap.String("image", c.image), zap.String("err", err.Error()))
		fmt.Fprintf(c.writer, output.ColorFmt(output.ANSIError, "-> could not pull image: %s", "\n"), err.Error())
		return err
	}
	defer pullResp.Close()
	HandleOutput(pullResp, secrets, c.writer)

	fmt.Fprintf(c.writer, output.ColorFmt(output.ANSIInfo, "-> pulled image: %s", "\n"), c.image)
	return nil
}

func (c *Container) Create() error {
	fmt.Fprintf(c.writer, output.ColorFmt(output.ANSIInfo, "-> %s created", "\n"), GetContainerName(c.name))

	c.containerConfig.Env = respectProxyEnv(c.containerConfig.Env)

	createResp, err := dockerClient.ContainerCreate(
		context.Background(),
		c.containerConfig,
		c.hostConfig,
		c.networkConfig,
		GetContainerName(c.name),
	)
	if err != nil {
		logging.GetLogger().Error("could not create container", zap.String("err", err.Error()))
		return err
	}
	c.containerID = createResp.ID
	return nil
}

func (c *Container) Run(wg *sync.WaitGroup, secrets []string, firstStoppedSvcCh chan string) error {
	defer func() { firstStoppedSvcCh <- c.name }()
	defer wg.Done()
	c.mutex.Lock()
	c.running = true
	fmt.Fprintf(c.writer, output.ColorFmt(output.ANSIInfo, "-> %s running", "\n"), GetContainerName(c.name))
	err := dockerClient.ContainerStart(
		context.Background(),
		c.containerID,
		types.ContainerStartOptions{},
	)
	if err != nil {
		logging.GetLogger().Error(
			"could not start container",
			zap.String("err", err.Error()),
			zap.String("containerID", c.containerID),
		)
		return err
	}
	logsResp, err := dockerClient.ContainerLogs(
		context.Background(),
		c.containerID,
		types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true},
	)
	if err != nil {
		logging.GetLogger().Error(
			"could not get container logs",
			zap.String("err", err.Error()),
			zap.String("containerID", c.containerID),
		)
		return err
	}
	defer logsResp.Close()
	c.mutex.Unlock()
	HandleOutput(logsResp, secrets, c.writer)

	return nil
}

func (c *Container) Stop() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.running {
		c.running = false
		stopTimeout, _ := time.ParseDuration("1s")
		err := dockerClient.ContainerStop(
			context.Background(),
			c.containerID,
			&stopTimeout,
		)
		if err != nil {
			logging.GetLogger().Error(
				"could not stop container",
				zap.String("err", err.Error()),
				zap.String("containerID", c.containerID),
			)
			return err
		}

		container, err := dockerClient.ContainerInspect(context.Background(), c.containerID)
		if err != nil {
			logging.GetLogger().Error(
				"could not inspect container",
				zap.String("err", err.Error()),
				zap.String("containerID", c.containerID),
			)
			return err
		}

		c.exitCode = container.State.ExitCode
		fmt.Fprintf(c.writer,
			output.ColorFmt(output.ANSIInfo, "-> %s container exited: %d", "\n"),
			GetContainerName(c.name),
			c.exitCode,
		)

		err = dockerClient.ContainerRemove(
			context.Background(),
			c.containerID,
			types.ContainerRemoveOptions{RemoveVolumes: true},
		)
		if err != nil {
			logging.GetLogger().Error(
				"could not remove container",
				zap.String("err", err.Error()),
				zap.String("containerID", c.containerID),
			)
			return err
		}
		fmt.Fprintf(c.writer, output.ColorFmt(output.ANSIInfo, "-> %s removed", "\n"), GetContainerName(c.name))

	}
	return nil
}

// TODO: clean up

func GetContainerName(serviceName string) string {
	return fmt.Sprintf(
		"vci-%s",
		serviceName,
	)
}

func GetImageName(serviceName string) string {
	return fmt.Sprintf(
		"vci-%s",
		serviceName,
	)
}

// func getAuthToken(image string, dockerRegistries []DockerRegistry) string {
func getAuthToken(image string, addressAuthTokens map[string]string) string {
	tagParts := strings.Split(image, "/")
	registry := tagParts[0]
	if strings.Contains(registry, ".") {
		// private
		for address, token := range addressAuthTokens {
			if address == registry {
				return token
			}
		}
	} else {
		for address, token := range addressAuthTokens {
			if strings.Contains(address, "https://registry.hub.docker.com") || strings.Contains(address, "https://index.docker.io") {
				return token
			}
		}
	}

	return ""
}

func respectProxyEnv(env []string) []string {
	config := httpproxy.FromEnvironment()
	if len(config.HTTPProxy) > 1 {
		env = append(env, fmt.Sprintf("HTTP_PROXY=%s", config.HTTPProxy))
		env = append(env, fmt.Sprintf("http_proxy=%s", config.HTTPProxy))
	}
	if len(config.HTTPSProxy) > 1 {
		env = append(env, fmt.Sprintf("HTTPS_PROXY=%s", config.HTTPSProxy))
		env = append(env, fmt.Sprintf("https_proxy=%s", config.HTTPSProxy))
	}
	if len(config.NoProxy) > 1 {
		env = append(env, fmt.Sprintf("NO_PROXY=%s", config.NoProxy))
		env = append(env, fmt.Sprintf("no_proxy=%s", config.NoProxy))
	}

	return env
}

func resolvePullImage(image string) string {
	parts := strings.Split(image, "/")

	if strings.Contains(parts[0], ".") {
		return image
	}

	if len(parts) > 0 {
		return fmt.Sprintf("docker.io/%s", image)
	}

	return fmt.Sprintf("docker.io/library/%s", image)
}

func findImageLocally(imageName string, cli *client.Client, ctx context.Context) error {
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		logging.GetLogger().Error("could not find image", zap.String("err", err.Error()))
		return err
	}
	for _, i := range images {
		for _, t := range i.RepoTags {
			if t == imageName {
				return nil
			}
		}
	}
	return fmt.Errorf("could not find image: %s", imageName)
}
