package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
	"go.uber.org/zap"
)

func NewImagePusher() *ImagePusher {
	return &ImagePusher{
		running: false,
	}
}

type ImagePusher struct {
	running bool

	response io.ReadCloser
}

func (iP *ImagePusher) Push(
	writer io.Writer,
	secrets []string,
	tag string,
	addressAuthTokens map[string]string,
) error {
	// Determine correct authToken
	authToken := getAuthToken(tag, addressAuthTokens)
	logging.GetLogger().Debug("pushing image",
		zap.String("tag", tag),
		zap.String("registry auth", authToken),
	)
	reader, err := dockerClient.ImagePush(context.Background(), tag, types.ImagePushOptions{
		All:          true,
		RegistryAuth: authToken,
	})
	if err != nil {
		return err
	}
	HandleOutput(reader, secrets, writer)
	if !iP.running {
		return fmt.Errorf("image push interrupted")
	}
	iP.GracefulStop()
	fmt.Fprintf(writer, output.ColorFmt(output.ANSIInfo, "-> pushed: %s", "\n"), tag)
	logging.GetLogger().Debug("finished pushing image", zap.String("tag", tag))

	return nil
}

func (iP *ImagePusher) GracefulStop() {
	if iP.running {
		iP.response.Close()
		iP.running = false
	}
}

func PushImage(
	writer io.Writer,
	tag string,
	addressAuthTokens map[string]string,
	secrets []string,
) error {
	// Determine correct authToken
	authToken := getAuthToken(tag, addressAuthTokens)
	logging.GetLogger().Debug("pushing image",
		zap.String("tag", tag),
		zap.String("registry auth", authToken),
	)
	reader, err := dockerClient.ImagePush(context.Background(), tag, types.ImagePushOptions{
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
