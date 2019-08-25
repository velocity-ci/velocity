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

// NewImagePusher returns a new Docker image pusher
func NewImagePusher() *ImagePusher {
	return &ImagePusher{
		running: false,
	}
}

// Represents a stoppable Docker image pusher
type ImagePusher struct {
	running bool

	response io.ReadCloser
}

// Push pushes a docker image with the given parameters
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
	iP.Stop()
	fmt.Fprintf(writer, output.ColorFmt(output.ANSIInfo, "-> pushed: %s", "\n"), tag)
	logging.GetLogger().Debug("finished pushing image", zap.String("tag", tag))

	return nil
}

// Stop interrupts an image push process
func (iP *ImagePusher) Stop() {
	if iP.running {
		iP.response.Close()
		iP.running = false
	}
}
