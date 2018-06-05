package velocity

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/network"
	"github.com/golang/glog"

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
	build *dockerComposeServiceBuild,
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
	build           *dockerComposeServiceBuild
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
		"vci-%s-%x",
		serviceName,
		md5.Sum([]byte(serviceName)),
	)
}

func getImageName(serviceName string) string {
	return fmt.Sprintf(
		"vci-%s-%x",
		serviceName,
		md5.Sum([]byte(serviceName)),
	)
}

func getAuthConfigsMap(dockerRegistries []DockerRegistry) map[string]types.AuthConfig {
	authConfigs := map[string]types.AuthConfig{}
	for _, r := range dockerRegistries {
		jsonAuthConfig, err := base64.URLEncoding.DecodeString(r.AuthorizationToken)
		if err != nil {
			glog.Error(err)
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
			glog.Errorf("build image err: %s", err)
		}
		sR.image = getImageName(sR.name)
		sR.containerConfig.Image = getImageName(sR.name)
	} else {
		// check if image exists locally before pulling
		if findImageLocally(sR.image, sR.dockerCli, sR.context) != nil {
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
				glog.Errorf("pull image err: %s", err)
			}
			defer pullResp.Close()
			handleOutput(pullResp, sR.params, sR.writer)

			sR.writer.Write([]byte(fmt.Sprintf("Pulled: %s", sR.image)))
		} else {
			sR.writer.Write([]byte(fmt.Sprintf("Got locally: %s", sR.image)))
		}

	}
}

func findImageLocally(imageName string, cli *client.Client, ctx context.Context) error {
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		glog.Errorf("image find err: %s", err)
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

func (sR *serviceRunner) Create() {
	sR.writer.Write([]byte(fmt.Sprintf("Creating container: %s", getContainerName(sR.name))))
	createResp, err := sR.dockerCli.ContainerCreate(
		sR.context,
		sR.containerConfig,
		sR.hostConfig,
		sR.networkConfig,
		getContainerName(sR.name),
	)
	if err != nil {
		glog.Errorf("container create err: %s", err)
	}
	sR.containerID = createResp.ID
}

func (sR *serviceRunner) Run(stop chan string) {
	sR.writer.Write([]byte(fmt.Sprintf("Running container: %s (%s)", getContainerName(sR.name), sR.containerID)))
	err := sR.dockerCli.ContainerStart(
		sR.context,
		sR.containerID,
		types.ContainerStartOptions{},
	)
	if err != nil {
		glog.Errorf("container %s start err: %s", sR.containerID, err)
	}
	logsResp, err := sR.dockerCli.ContainerLogs(
		sR.context,
		sR.containerID,
		types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true},
	)
	if err != nil {
		glog.Errorf("container %s logs err: %s", sR.containerID, err)
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
		glog.Errorf("container %s stop err: %s", sR.containerID, err)
	}

	container, err := sR.dockerCli.ContainerInspect(sR.context, sR.containerID)
	if err != nil {
		glog.Errorf("container %s inspect err: %s", sR.containerID, err)
	}

	sR.exitCode = container.State.ExitCode
	sR.writer.Write([]byte(fmt.Sprintf("Container %s exited: %d", sR.containerID, sR.exitCode)))
	sR.writer.Write([]byte(fmt.Sprintf("container %s status: %s", sR.containerID, container.State.Status)))

	if !sR.removing {
		sR.removing = true
		err = sR.dockerCli.ContainerRemove(
			sR.context,
			sR.containerID,
			types.ContainerRemoveOptions{RemoveVolumes: true},
		)
		if err != nil {
			glog.Errorf("container %s remove err: %s", sR.containerID, err)
		}
		sR.writer.Write([]byte(fmt.Sprintf("Removed container: %s (%s)", getContainerName(sR.name), sR.containerID)))
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
	cwd, _ := os.Getwd()
	buildContext = fmt.Sprintf("%s/%s", cwd, buildContext)

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
	handleOutput(buildResp.Body, parameters, writer)

	return nil
}

func handleOutput(body io.ReadCloser, parameters map[string]Parameter, writer io.Writer) {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		allBytes := scanner.Bytes()

		o := ""
		if strings.Contains(string(allBytes), "status") {
			o = handlePullPushOutput(allBytes)
		} else if strings.Contains(string(allBytes), "stream") {
			o = handleBuildOutput(allBytes)
		} else if strings.Contains(string(allBytes), "progressDetail") {
			o = "*"
		} else {
			o = handleLogOutput(allBytes)
		}

		if o != "*" {
			glog.Infof(o)
			for _, p := range parameters {
				if p.IsSecret {
					o = strings.Replace(o, p.Value, "***", -1)
				}
			}
			writer.Write([]byte(o))
		}
	}
}

func handleLogOutput(b []byte) string {
	if len(b) <= 8 {
		return ""
	}
	return string(b[8:])
}

var imageIDProgress = map[string]string{}

func handlePullPushOutput(b []byte) string {
	type pullOutput struct {
		Status   string `json:"status"`
		Progress string `json:"progress"`
		ID       string `json:"id"`
	}
	var o pullOutput
	json.Unmarshal(b, &o)

	s := ""
	if len(o.ID) > 0 {
		s += fmt.Sprintf("%s: ", o.ID)
	}
	if len(o.Progress) > 0 {
		s += o.Progress
	} else {
		s += o.Status
	}
	// add padding to 80
	for len(s) < 80 {
		s += " "
	}
	if strings.Contains(o.Status, "Downloaded newer image") ||
		strings.Contains(o.Status, "Pulling from") ||
		strings.Contains(o.Status, "Pull complete") {
		return s
	}

	return fmt.Sprintf("%s\r", s)
}

func handleBuildOutput(b []byte) string {
	type buildOutput struct {
		Stream string `json:"stream"`
	}
	var o buildOutput
	json.Unmarshal(b, &o)
	return strings.TrimSpace(o.Stream)
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
