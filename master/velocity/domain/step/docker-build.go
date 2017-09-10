package step

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
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
	return fmt.Sprintf("dockerfile: %s, context: %s, tags: %s")
}

func (dB *DockerBuild) Execute() error {
	cli, err := client.NewEnvClient()
	ctx := context.Background()
	if err != nil {
		panic(err)
	}

	dB.Context = fmt.Sprintf("%s/%s", os.Getwd(), dB.Context)

	excludes, err := build.ReadDockerignore(dB.Context)
	if err != nil {
		return err
	}
	buildCtx, err := archive.TarWithOptions(dB.Context, &archive.TarOptions{
		ExcludePatterns: excludes,
	})

	in, out, err := term.StdStreams()
	dockerCli := command.NewDockerCli(in, out, err)

	res, err := dockerCli.Client().ImageBuild(ctx, buildCtx, types.ImageBuildOptions{
		Dockerfile: dB.Dockerfile,
		Tags:       dB.Tags,
	})

	aux := func(auxJSON *json.RawMessage) {
		var result types.BuildResult
		if err := json.Unmarshal(*auxJSON, &result); err != nil {
			fmt.Fprintf(dockerCli.Err(), "Failed to parse aux message: %s", err)
		} else {
			imageID = result.ID
		}
	}

	jsonmessage.DisplayJSONMessagesStream(res.Body, dockerCli.Out(), dockerCli.Out().FD(), dockerCli.Out().IsTerminal(), aux)
	return nil
}

func (dB *DockerBuild) Validate(params []domain.Parameter) error {
	return nil
}

func (dB *DockerBuild) SetParams(params []domain.Parameter) error {
	return nil
}
