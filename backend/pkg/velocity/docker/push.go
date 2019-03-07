package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"

	"github.com/docker/docker/client"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
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
	reader, err := cli.ImagePush(ctx, tag, types.ImagePushOptions{
		All:          true,
		RegistryAuth: authToken,
	})
	out.HandleOutput(reader, secrets, writer)
	if err != nil {
		return err
	}
	fmt.Fprintf(writer, out.ColorFmt(out.ANSIInfo, "-> pushed: %s"), tag)

	return nil
}
