package docker

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
	"go.uber.org/zap"
)

func BuildContainer(
	buildContext string,
	dockerfile string,
	tags []string,
	secrets []string,
	writer io.Writer,
	authConfigs map[string]types.AuthConfig,
) error {
	logging.GetLogger().Debug("building image", zap.String("Dockerfile", dockerfile), zap.String("build context", buildContext))

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
	out.HandleOutput(buildResp.Body, secrets, writer)

	logging.GetLogger().Debug("finished building image", zap.String("Dockerfile", dockerfile), zap.String("build context", buildContext))
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
