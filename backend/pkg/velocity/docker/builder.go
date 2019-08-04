package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
	"go.uber.org/zap"
)

func NewImageBuilder() *ImageBuilder {
	return &ImageBuilder{
		running: false,
	}
}

type ImageBuilder struct {
	running bool

	buildResp types.ImageBuildResponse
}

func (iB *ImageBuilder) IsRunning() bool {
	return iB.running
}

func (iB *ImageBuilder) Build(
	writer io.Writer,
	secrets []string,
	buildContext string,
	dockerfile string,
	tags []string,
	authConfigs map[string]types.AuthConfig,
) error {
	logging.GetLogger().Debug("building image",
		zap.String("Dockerfile", dockerfile),
		zap.String("build context", buildContext),
		zap.Strings("tags", tags),
	)

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

	iB.buildResp, err = dockerClient.ImageBuild(context.Background(), buildCtx, types.ImageBuildOptions{
		AuthConfigs: authConfigs,
		PullParent:  true,
		Remove:      true,
		Dockerfile:  dockerfile,
		Tags:        tags,
	})
	if err != nil {
		return err
	}
	iB.running = true
	defer iB.buildResp.Body.Close()
	HandleOutput(iB.buildResp.Body, secrets, writer)
	if !iB.running {
		return fmt.Errorf("image build interrupted")
	}
	iB.Stop()
	fmt.Fprintf(writer, output.ColorFmt(output.ANSIInfo, "-> built: %s", "\n"), strings.Join(tags, ", "))
	logging.GetLogger().Debug("finished building image", zap.String("Dockerfile", dockerfile), zap.String("build context", buildContext))
	return nil
}

func (iB *ImageBuilder) Stop() error {
	if iB.running {
		iB.buildResp.Body.Close()
		iB.running = false
	}
	return nil
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
