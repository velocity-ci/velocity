package velocity

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
)

func runContainer(
	pullImage string,
	config *container.Config,
	hostConfig *container.HostConfig,
	parameters map[string]Parameter,
	writer io.Writer,
) (int, error) {

	cli, err := client.NewEnvClient()
	if err != nil {
		return 1, err
	}

	ctx := context.Background()

	pullResp, err := cli.ImagePull(ctx, pullImage, types.ImagePullOptions{})
	if err != nil {
		return 1, err
	}
	defer pullResp.Close()
	handleOutput(pullResp, parameters, writer)

	createResp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, "")
	if err != nil {
		return 1, err
	}

	err = cli.ContainerStart(ctx, createResp.ID, types.ContainerStartOptions{})
	if err != nil {
		return 1, err
	}

	logsResp, err := cli.ContainerLogs(ctx, createResp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		return 1, err
	}
	defer logsResp.Close()
	handleOutput(logsResp, parameters, writer)

	container, err := cli.ContainerInspect(ctx, createResp.ID)
	if err != nil {
		return 1, err
	}
	stopTimeout, _ := time.ParseDuration("30s")
	err = cli.ContainerStop(ctx, createResp.ID, &stopTimeout)
	if err != nil {
		return 1, err
	}
	err = cli.ContainerRemove(ctx, createResp.ID, types.ContainerRemoveOptions{RemoveVolumes: true})
	if err != nil {
		return 1, err
	}

	return container.State.ExitCode, nil
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
