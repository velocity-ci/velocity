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

func NewServiceRunner(
	cli *client.Client,
	ctx context.Context,
	writer io.Writer,
	wg *sync.WaitGroup,
	secrets []string,
	name string,
	image string,
	build *v3.DockerComposeServiceBuild,
	containerConfig *container.Config,
	hostConfig *container.HostConfig,
	networkConfig *network.NetworkingConfig,
	networkID string,
) *ServiceRunner {
	return &ServiceRunner{
		dockerCli:       cli,
		context:         ctx,
		writer:          writer,
		wg:              wg,
		secrets:         secrets,
		name:            name,
		image:           image,
		build:           build,
		containerConfig: containerConfig,
		hostConfig:      hostConfig,
		networkConfig:   networkConfig,
		networkID:       networkID,
		stopped:         false,
	}
}

type ServiceRunner struct {
	dockerCli *client.Client
	context   context.Context
	writer    io.Writer

	name            string
	image           string
	build           *v3.DockerComposeServiceBuild
	containerConfig *container.Config
	hostConfig      *container.HostConfig
	networkConfig   *network.NetworkingConfig

	networkID   string
	containerID string
	ExitCode    int
	stopped     bool

	wg      *sync.WaitGroup
	secrets []string
}

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

func (sR *ServiceRunner) PullOrBuild(authConfigs map[string]types.AuthConfig, addressAuthToken map[string]string) {
	if sR.build != nil && (sR.build.Dockerfile != "" || sR.build.Context != "") {
		// authConfigs := getAuthConfigsMap(dockerRegistries)
		builder := NewImageBuilder(sR.dockerCli, sR.context, sR.writer, sR.secrets)
		err := builder.Build(
			sR.build.Context,
			sR.build.Dockerfile,
			[]string{GetImageName(sR.name)},
			authConfigs,
		)
		if err != nil {
			logging.GetLogger().Error("could not build image", zap.String("err", err.Error()))
		}
		sR.image = GetImageName(sR.name)
		sR.containerConfig.Image = GetImageName(sR.name)
	} else {
		// check if image exists locally before pulling
		sR.image = resolvePullImage(sR.image)
		sR.containerConfig.Image = resolvePullImage(sR.image)
		authToken := getAuthToken(sR.image, addressAuthToken)
		pullResp, err := sR.dockerCli.ImagePull(
			sR.context,
			sR.image,
			types.ImagePullOptions{
				RegistryAuth: authToken,
			},
		)
		if err != nil {
			logging.GetLogger().Error("could not pull image", zap.String("image", sR.image), zap.String("err", err.Error()))
			fmt.Fprintf(sR.writer, output.ColorFmt(output.ANSIError, "-> could not pull image: %s", "\n"), err.Error())
			return
		}
		defer pullResp.Close()
		HandleOutput(pullResp, sR.secrets, sR.writer)

		fmt.Fprintf(sR.writer, output.ColorFmt(output.ANSIInfo, "-> pulled image: %s", "\n"), sR.image)
	}
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

func (sR *ServiceRunner) Create() {
	fmt.Fprintf(sR.writer, output.ColorFmt(output.ANSIInfo, "-> %s created", "\n"), GetContainerName(sR.name))

	sR.containerConfig.Env = respectProxyEnv(sR.containerConfig.Env)

	createResp, err := sR.dockerCli.ContainerCreate(
		sR.context,
		sR.containerConfig,
		sR.hostConfig,
		sR.networkConfig,
		GetContainerName(sR.name),
	)
	if err != nil {
		logging.GetLogger().Error("could not create container", zap.String("err", err.Error()))
	}
	sR.containerID = createResp.ID
}

func (sR *ServiceRunner) Run(stop chan string) { // rename to start
	fmt.Fprintf(sR.writer, output.ColorFmt(output.ANSIInfo, "-> %s running", "\n"), GetContainerName(sR.name))
	err := sR.dockerCli.ContainerStart(
		sR.context,
		sR.containerID,
		types.ContainerStartOptions{},
	)
	if err != nil {
		logging.GetLogger().Error("could not start container", zap.String("err", err.Error()), zap.String("containerID", sR.containerID))
	}
	logsResp, err := sR.dockerCli.ContainerLogs(
		sR.context,
		sR.containerID,
		types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true},
	)
	if err != nil {
		logging.GetLogger().Error("could not get container logs", zap.String("err", err.Error()), zap.String("containerID", sR.containerID))
	}
	defer logsResp.Close()
	HandleOutput(logsResp, sR.secrets, sR.writer)

	stop <- sR.name
}

func (sR *ServiceRunner) Stop() {
	if !sR.stopped {
		sR.stopped = true
		defer sR.wg.Done()

		stopTimeout, _ := time.ParseDuration("3s")
		err := sR.dockerCli.ContainerStop(
			sR.context,
			sR.containerID,
			&stopTimeout,
		)
		if err != nil {
			logging.GetLogger().Error("could not stop container", zap.String("err", err.Error()), zap.String("containerID", sR.containerID))
		}

		container, err := sR.dockerCli.ContainerInspect(sR.context, sR.containerID)
		if err != nil {
			logging.GetLogger().Error("could not inspect container", zap.String("err", err.Error()), zap.String("containerID", sR.containerID))
		}

		sR.ExitCode = container.State.ExitCode
		fmt.Fprintf(sR.writer, output.ColorFmt(output.ANSIInfo, "-> %s container exited: %d", "\n"), GetContainerName(sR.name), sR.ExitCode)

		err = sR.dockerCli.ContainerRemove(
			sR.context,
			sR.containerID,
			types.ContainerRemoveOptions{RemoveVolumes: true},
		)
		if err != nil {
			logging.GetLogger().Error("could not remove container", zap.String("err", err.Error()), zap.String("containerID", sR.containerID))
		}
		fmt.Fprintf(sR.writer, output.ColorFmt(output.ANSIInfo, "-> %s removed", "\n"), GetContainerName(sR.name))

	}
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
