package task

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/image/build"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/term"
)

func runContainer(
	pullImage string,
	config *container.Config,
	hostConfig *container.HostConfig,
	parameters map[string]Parameter,
	emitter Emitter,
) (int, error) {
	ctx := context.Background()
	stdIn, stdOut, stdErr := term.StdStreams()
	dockerCli := command.NewDockerCli(stdIn, stdOut, stdErr)
	dockerCli.Initialize(cliflags.NewClientOptions())

	pullResponse, pullErr := dockerCli.Client().ImagePull(ctx, pullImage, types.ImagePullOptions{})
	if pullErr == nil {
		defer pullResponse.Close()
		handleOutput(pullResponse, parameters, emitter)
	}

	createResponse, err := dockerCli.Client().ContainerCreate(ctx, config, hostConfig, nil, "")
	if err != nil {
		return -1, err
	}

	dockerCli.Client().ContainerStart(ctx, createResponse.ID, types.ContainerStartOptions{})
	logsResponse, err := dockerCli.Client().ContainerLogs(ctx, createResponse.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	defer logsResponse.Close()
	if err != nil {
		return -1, err
	}
	handleOutput(logsResponse, parameters, emitter)

	c, _ := dockerCli.Client().ContainerInspect(ctx, createResponse.ID)
	d, _ := time.ParseDuration("1m")
	err = dockerCli.Client().ContainerStop(ctx, createResponse.ID, &d)
	if err != nil {
		fmt.Println(err)
	}

	err = dockerCli.Client().ContainerRemove(ctx, createResponse.ID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
	})
	if err != nil {
		fmt.Println(err)
	}

	return c.State.ExitCode, nil
}

func buildContainer(buildContext string, dockerfile string, tags []string, parameters map[string]Parameter, emitter Emitter) error {
	ctx := context.Background()

	cwd, _ := os.Getwd()
	buildContext = fmt.Sprintf("%s/%s", cwd, buildContext)

	excludes, err := build.ReadDockerignore(buildContext)
	if err != nil {
		emitter.SetStatus("failed")
		emitter.Write([]byte(err.Error()))
		return err
	}
	buildCtx, err := archive.TarWithOptions(buildContext, &archive.TarOptions{
		ExcludePatterns: excludes,
	})

	stdIn, stdOut, stdErr := term.StdStreams()
	dockerCli := command.NewDockerCli(stdIn, stdOut, stdErr)
	dockerCli.Initialize(cliflags.NewClientOptions())

	buildResponse, err := dockerCli.Client().ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		PullParent: true,
		Remove:     true,
		Dockerfile: dockerfile,
		Tags:       tags,
	})
	if err != nil {
		emitter.SetStatus("failed")
		emitter.Write([]byte(err.Error()))
		return err
	}
	defer buildResponse.Body.Close()

	handleOutput(buildResponse.Body, parameters, emitter)

	return nil
}

func handleOutput(body io.ReadCloser, parameters map[string]Parameter, emitter Emitter) {
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
			emitter.Write([]byte(o))
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
