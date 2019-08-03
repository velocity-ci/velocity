package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
	"go.uber.org/zap"

	"github.com/docker/docker/client"
)

func NewImagePusher(
	cli *client.Client,
	ctx context.Context,
	writer io.Writer,
	secrets []string,
) *ImagePusher {
	return &ImagePusher{
		dockerCli: cli,
		context:   ctx,
		writer:    writer,
		secrets:   secrets,
		stopped:   false,
	}
}

type ImagePusher struct {
	dockerCli *client.Client
	context   context.Context
	writer    io.Writer
	stopped   bool
	secrets   []string

	response io.ReadCloser
}

func (iP *ImagePusher) Push(
	tag string,
	addressAuthTokens map[string]string,
) error {
	// Determine correct authToken
	authToken := getAuthToken(tag, addressAuthTokens)
	logging.GetLogger().Debug("pushing image",
		zap.String("tag", tag),
		zap.String("registry auth", authToken),
	)
	reader, err := iP.dockerCli.ImagePush(iP.context, tag, types.ImagePushOptions{
		All:          true,
		RegistryAuth: authToken,
	})
	if err != nil {
		return err
	}
	HandleOutput(reader, iP.secrets, iP.writer)
	if iP.stopped {
		return fmt.Errorf("image push interrupted")
	}
	iP.stopped = true
	fmt.Fprintf(iP.writer, output.ColorFmt(output.ANSIInfo, "-> pushed: %s", "\n"), tag)
	logging.GetLogger().Debug("finished pushing image", zap.String("tag", tag))

	return nil
}

func (iP *ImagePusher) GracefulStop() {
	if !iP.stopped {
		iP.response.Close()
		iP.stopped = true
	}
}

func PushImage(
	writer io.Writer,
	tag string,
	addressAuthTokens map[string]string,
	secrets []string,
) error {
	cli, _ := client.NewEnvClient()
	ctx := context.Background()
	// Determine correct authToken
	authToken := getAuthToken(tag, addressAuthTokens)
	logging.GetLogger().Debug("pushing image",
		zap.String("tag", tag),
		zap.String("registry auth", authToken),
	)
	reader, err := cli.ImagePush(ctx, tag, types.ImagePushOptions{
		All:          true,
		RegistryAuth: authToken,
	})
	if err != nil {
		return err
	}
	HandleOutput(reader, secrets, writer)
	fmt.Fprintf(writer, output.ColorFmt(output.ANSIInfo, "-> pushed: %s", "\n"), tag)

	return nil
}
