package velocity

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
		networkID:       networkID,
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

	networkID   string
	containerID string
	exitCode    int

	wg     *sync.WaitGroup
	params map[string]Parameter

	allServices []*serviceRunner
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

func (sR *serviceRunner) setServices(s []*serviceRunner) {
	sR.allServices = s
}

func (sR *serviceRunner) PullOrBuild() {
	if sR.build != nil {
		err := buildContainer(
			sR.build.Context,
			sR.build.Dockerfile,
			[]string{getImageName(sR.name)},
			sR.params,
			sR.writer,
		)
		if err != nil {
			log.Println(err)
		}
		sR.image = getImageName(sR.name)
	} else {
		// check if image exists locally before pulling
		if findImageLocally(sR.image, sR.dockerCli, sR.context) != nil {
			sR.image = resolvePullImage(sR.image)

			pullResp, err := sR.dockerCli.ImagePull(
				sR.context,
				sR.image,
				types.ImagePullOptions{},
			)
			if err != nil {
				log.Println(err)
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
		log.Println(err)
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
		nil,
		getContainerName(sR.name),
	)
	if err != nil {
		log.Println(err)
	}
	sR.containerID = createResp.ID
}

func (sR *serviceRunner) Run() {
	sR.writer.Write([]byte(fmt.Sprintf("Running container: %s", getContainerName(sR.name))))
	err := sR.dockerCli.ContainerStart(
		sR.context,
		sR.containerID,
		types.ContainerStartOptions{},
	)
	if err != nil {
		log.Println(err)
	}
	logsResp, err := sR.dockerCli.ContainerLogs(
		sR.context,
		sR.containerID,
		types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true},
	)
	if err != nil {
		log.Println(err)
	}
	defer logsResp.Close()
	handleOutput(logsResp, sR.params, sR.writer)

	for _, s := range sR.allServices {
		go s.Stop()
	}
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
		log.Println(err)
	}

	container, err := sR.dockerCli.ContainerInspect(sR.context, sR.containerID)
	if err != nil {
		log.Println(err)
	}

	sR.exitCode = container.State.ExitCode
	sR.writer.Write([]byte(fmt.Sprintf("Container %s exited: %d", getContainerName(sR.name), sR.exitCode)))

	err = sR.dockerCli.ContainerRemove(
		sR.context,
		sR.containerID,
		types.ContainerRemoveOptions{RemoveVolumes: true},
	)
	if err != nil {
		log.Println(err)
	}
}

func buildContainer(
	buildContext string,
	dockerfile string,
	tags []string,
	parameters map[string]Parameter,
	writer io.Writer,
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
		PullParent: true,
		Remove:     true,
		Dockerfile: dockerfile,
		Tags:       tags,
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
			o = handlePullOutput(allBytes)
		} else if strings.Contains(string(allBytes), "stream") {
			o = handleBuildOutput(allBytes)
		} else {
			o = handleLogOutput(allBytes)
		}

		if o != "*" {
			for _, p := range parameters {
				if p.Secret {
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

func handlePullOutput(b []byte) string {
	type pullOutput struct {
		Status   string `json:"status"`
		Progress string `json:"progress"`
		ID       string `json:"id"`
	}
	var o pullOutput
	json.Unmarshal(b, &o)

	if strings.Contains(o.Status, "Pulling") ||
		strings.Contains(o.Status, "Download complete") ||
		strings.Contains(o.Status, "Digest") {
		return fmt.Sprintf("%s: %s",
			o.ID,
			strings.TrimSpace(o.Status),
		)
	}

	return "*"
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
