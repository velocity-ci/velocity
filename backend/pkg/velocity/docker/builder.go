package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
	"go.uber.org/zap"
)

func NewImageBuilder(
	cli *client.Client,
	ctx context.Context,
	writer io.Writer,
	secrets []string,
) *ImageBuilder {
	return &ImageBuilder{
		dockerCli: cli,
		context:   ctx,
		writer:    writer,
		secrets:   secrets,
		stopped:   false,
	}
}

type ImageBuilder struct {
	dockerCli *client.Client
	context   context.Context
	writer    io.Writer
	stopped   bool
	secrets   []string

	buildResp types.ImageBuildResponse
}

func (iB *ImageBuilder) Build(
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

	iB.buildResp, err = iB.dockerCli.ImageBuild(iB.context, buildCtx, types.ImageBuildOptions{
		AuthConfigs: authConfigs,
		PullParent:  true,
		Remove:      true,
		Dockerfile:  dockerfile,
		Tags:        tags,
	})
	if err != nil {
		return err
	}

	defer iB.buildResp.Body.Close()
	HandleOutput(iB.buildResp.Body, iB.secrets, iB.writer)
	if iB.stopped {
		return fmt.Errorf("image build interrupted")
	}
	iB.stopped = true
	fmt.Fprintf(iB.writer, output.ColorFmt(output.ANSIInfo, "-> built: %s", "\n"), tags)
	logging.GetLogger().Debug("finished building image", zap.String("Dockerfile", dockerfile), zap.String("build context", buildContext))
	return nil
}

func (iB *ImageBuilder) GracefulStop() {
	if !iB.stopped {
		iB.buildResp.Body.Close()
		iB.stopped = true
	}
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
