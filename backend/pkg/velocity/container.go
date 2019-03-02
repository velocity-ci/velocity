package velocity

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	vio "github.com/velocity-ci/velocity/backend/pkg/velocity/io"
	dockercompose "github.com/velocity-ci/velocity/backend/pkg/velocity/step/docker/compose/v3"

	"github.com/docker/docker/api/types/network"
	"go.uber.org/zap"
	"golang.org/x/net/http/httpproxy"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

func newServiceRunner(
	cli *client.Client,
	ctx context.Context,
	writer io.Writer,
	wg *sync.WaitGroup,
	params map[string]Parameter,
	name string,
	image string,
	build *dockercompose.DockerComposeServiceBuild,
	containerConfig *container.Config,
	hostConfig *container.HostConfig,
	networkConfig *network.NetworkingConfig,
	networkID string,
) *serviceRunner {
	return &serviceRunner{
		dockerCli:       cli,
		context:         ctx,
		writer:          writer,
		wg:              wg,
		params:          params,
		name:            name,
		image:           image,
		build:           build,
		containerConfig: containerConfig,
		hostConfig:      hostConfig,
		networkConfig:   networkConfig,
		networkID:       networkID,
		removing:        false,
	}
}

type serviceRunner struct {
	dockerCli *client.Client
	context   context.Context
	writer    io.Writer

	name            string
	image           string
	build           *dockercompose.DockerComposeServiceBuild
	containerConfig *container.Config
	hostConfig      *container.HostConfig
	networkConfig   *network.NetworkingConfig

	networkID   string
	containerID string
	exitCode    int
	removing    bool

	wg     *sync.WaitGroup
	params map[string]Parameter
}

func getContainerName(serviceName string) string {
	return fmt.Sprintf(
		"vci-%s",
		serviceName,
	)
}

func getImageName(serviceName string) string {
	return fmt.Sprintf(
		"vci-%s",
		serviceName,
	)
}

func getAuthConfigsMap(dockerRegistries []DockerRegistry) map[string]types.AuthConfig {
	authConfigs := map[string]types.AuthConfig{}
	for _, r := range dockerRegistries {
		jsonAuthConfig, err := base64.URLEncoding.DecodeString(r.AuthorizationToken)
		if err != nil {
			GetLogger().Error(
				"could not decode registry auth config",
				zap.String("err", err.Error()),
				zap.String("registry", r.Address),
			)

		}
		var authConfig types.AuthConfig
		err = json.Unmarshal(jsonAuthConfig, &authConfig)
		authConfigs[r.Address] = authConfig
	}

	return authConfigs
}

func getAuthToken(image string, dockerRegistries []DockerRegistry) string {
	tagParts := strings.Split(image, "/")
	registry := tagParts[0]
	if strings.Contains(registry, ".") {
		// private
		for _, r := range dockerRegistries {
			if r.Address == registry {
				return r.AuthorizationToken
			}
		}
	} else {
		for _, r := range dockerRegistries {
			if strings.Contains(r.Address, "https://registry.hub.docker.com") || strings.Contains(r.Address, "https://index.docker.io") {
				return r.AuthorizationToken
			}
		}
	}

	return ""
}

func (sR *serviceRunner) PullOrBuild(dockerRegistries []DockerRegistry) {
	imageIDProgress = map[string]string{}
	if sR.build != nil && (sR.build.Dockerfile != "" || sR.build.Context != "") {
		authConfigs := getAuthConfigsMap(dockerRegistries)
		err := buildContainer(
			sR.build.Context,
			sR.build.Dockerfile,
			[]string{getImageName(sR.name)},
			sR.params,
			sR.writer,
			authConfigs,
		)
		if err != nil {
			GetLogger().Error("could not build image", zap.String("err", err.Error()))
		}
		sR.image = getImageName(sR.name)
		sR.containerConfig.Image = getImageName(sR.name)
	} else {
		// check if image exists locally before pulling
		sR.image = resolvePullImage(sR.image)
		sR.containerConfig.Image = resolvePullImage(sR.image)
		authToken := getAuthToken(sR.image, dockerRegistries)
		pullResp, err := sR.dockerCli.ImagePull(
			sR.context,
			sR.image,
			types.ImagePullOptions{
				RegistryAuth: authToken,
			},
		)
		if err != nil {
			GetLogger().Error("could not pull image", zap.String("err", err.Error()))
		}
		defer pullResp.Close()
		handleOutput(pullResp, sR.params, sR.writer)

		fmt.Fprintf(sR.writer, colorFmt(ansiInfo, "-> pulled image: %s"), sR.image)

	}
}

func findImageLocally(imageName string, cli *client.Client, ctx context.Context) error {
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		GetLogger().Error("could not find image", zap.String("err", err.Error()))
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

func (sR *serviceRunner) Create() {
	fmt.Fprintf(sR.writer, colorFmt(ansiInfo, "-> %s created"), getContainerName(sR.name))

	sR.containerConfig.Env = respectProxyEnv(sR.containerConfig.Env)

	createResp, err := sR.dockerCli.ContainerCreate(
		sR.context,
		sR.containerConfig,
		sR.hostConfig,
		sR.networkConfig,
		getContainerName(sR.name),
	)
	if err != nil {
		GetLogger().Error("could not create container", zap.String("err", err.Error()))
	}
	sR.containerID = createResp.ID
}

func (sR *serviceRunner) Run(stop chan string) {
	fmt.Fprintf(sR.writer, colorFmt(ansiInfo, "-> %s running"), getContainerName(sR.name))
	err := sR.dockerCli.ContainerStart(
		sR.context,
		sR.containerID,
		types.ContainerStartOptions{},
	)
	if err != nil {
		GetLogger().Error("could not start container", zap.String("err", err.Error()), zap.String("containerID", sR.containerID))
	}
	logsResp, err := sR.dockerCli.ContainerLogs(
		sR.context,
		sR.containerID,
		types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true},
	)
	if err != nil {
		GetLogger().Error("could not get container logs", zap.String("err", err.Error()), zap.String("containerID", sR.containerID))
	}
	defer logsResp.Close()
	handleOutput(logsResp, sR.params, sR.writer)

	stop <- sR.name
}

func (sR *serviceRunner) Stop() {
	defer sR.wg.Done()

	stopTimeout, _ := time.ParseDuration("30s")
	err := sR.dockerCli.ContainerStop(
		sR.context,
		sR.containerID,
		&stopTimeout,
	)
	if err != nil {
		GetLogger().Error("could not stop container", zap.String("err", err.Error()), zap.String("containerID", sR.containerID))
	}

	container, err := sR.dockerCli.ContainerInspect(sR.context, sR.containerID)
	if err != nil {
		GetLogger().Error("could not inspect container", zap.String("err", err.Error()), zap.String("containerID", sR.containerID))
	}

	sR.exitCode = container.State.ExitCode
	fmt.Fprintf(sR.writer, colorFmt(ansiInfo, "-> %s container exited: %d"), getContainerName(sR.name), sR.exitCode)

	if !sR.removing {
		sR.removing = true
		err = sR.dockerCli.ContainerRemove(
			sR.context,
			sR.containerID,
			types.ContainerRemoveOptions{RemoveVolumes: true},
		)
		if err != nil {
			GetLogger().Error("could not remove container", zap.String("err", err.Error()), zap.String("containerID", sR.containerID))
		}
		fmt.Fprintf(sR.writer, colorFmt(ansiInfo, "-> %s removed"), getContainerName(sR.name))

	}
}

func buildContainer(
	buildContext string,
	dockerfile string,
	tags []string,
	parameters map[string]Parameter,
	writer io.Writer,
	authConfigs map[string]types.AuthConfig,
) error {
	GetLogger().Debug("building image", zap.String("Dockerfile", dockerfile), zap.String("build context", buildContext))

	excludes, err := readDockerignore(buildContext)
	if err != nil {
		return err
	}

	buildCtx, err := archive.TarWithOptions(buildContext, &archive.TarOptions{
		ExcludePatterns: excludes,
	})

	if err != nil {
		return err
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	ctx := context.Background()

	buildResp, err := cli.ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		AuthConfigs: authConfigs,
		PullParent:  true,
		Remove:      true,
		Dockerfile:  dockerfile,
		Tags:        tags,
	})
	if err != nil {
		return err
	}

	defer buildResp.Body.Close()
	vio.HandleOutput(buildResp.Body, parameters, writer)

	GetLogger().Debug("finished building image", zap.String("Dockerfile", dockerfile), zap.String("build context", buildContext))
	return nil
}

func resolvePullImage(image string) string {
	parts := strings.Split(image, "/")

	if strings.Contains(parts[0], ".") {
		return image
	}

	return fmt.Sprintf("docker.io/%s", image)
}

// From: https://github.com/docker/cli/blob/c202b4b98704876b0476a8fda073c5ffa14ff76d/cli/command/image/build/dockerignore.go
// ReadDockerignore reads the .dockerignore file in the context directory and
// returns the list of paths to exclude
func readDockerignore(contextDir string) ([]string, error) {
	var excludes []string

	f, err := os.Open(filepath.Join(contextDir, ".dockerignore"))
	switch {
	case os.IsNotExist(err):
		return excludes, nil
	case err != nil:
		return nil, err
	}
	defer f.Close()

	return dockerignore.ReadAll(f)
}
