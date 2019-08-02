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
