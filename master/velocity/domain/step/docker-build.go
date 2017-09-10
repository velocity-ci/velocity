package step

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/image/build"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/velocity-ci/velocity/master/velocity/domain"
)

type DockerBuild struct {
	domain.BaseStep
	Dockerfile string   `json:"dockerfile" yaml:"dockerfile"`
	Context    string   `json:"context" yaml:"context"`
	Tags       []string `json:"tags" yaml:"tags"`
}

func (dB *DockerBuild) SetEmitter(e func(string)) {
	dB.Emit = e
}

func (dB DockerBuild) GetType() string {
	return "build"
}

func (dB DockerBuild) GetDescription() string {
	return dB.Description
}

func (dB DockerBuild) GetDetails() string {
	return fmt.Sprintf("dockerfile: %s, context: %s, tags: %s", dB.Dockerfile, dB.Context, dB.Tags)
}

func (dB *DockerBuild) Execute() error {
	ctx := context.Background()

	cwd, _ := os.Getwd()
	dB.Context = fmt.Sprintf("%s/%s", cwd, dB.Context)

	excludes, err := build.ReadDockerignore(dB.Context)
	if err != nil {
		return err
	}
	buildCtx, err := archive.TarWithOptions(dB.Context, &archive.TarOptions{
		ExcludePatterns: excludes,
	})

	stdIn, stdOut, stdErr := term.StdStreams()
	dockerCli := command.NewDockerCli(stdIn, stdOut, stdErr)
	dockerCli.Initialize(cliflags.NewClientOptions())

	res, err := dockerCli.Client().ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		Dockerfile: dB.Dockerfile,
		Tags:       dB.Tags,
	})

	imageID := ""
	aux := func(auxJSON *json.RawMessage) {
		var result types.BuildResult
		if err := json.Unmarshal(*auxJSON, &result); err != nil {
			fmt.Fprintf(dockerCli.Err(), "Failed to parse aux message: %s", err)
		} else {
			imageID = result.ID
		}
	}

	jsonmessage.DisplayJSONMessagesStream(res.Body, dockerCli.Out(), dockerCli.Out().FD(), false, aux)
	return nil
}

func (dB *DockerBuild) Validate(params []domain.Parameter) error {
	return nil
}

func (dB *DockerBuild) SetParams(params []domain.Parameter) error {
	return nil
}
