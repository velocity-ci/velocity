package docker

import (
	"os"
	"sync"

	"github.com/docker/docker/client"
)

var once sync.Once
var dockerClient *client.Client

func init() {
	// Minimum supported version as recommended by https://docs.docker.com/develop/sdk/#api-version-matrix
	os.Setenv("DOCKER_API_VERSION", "1.24")

	once.Do(func() {
		var err error
		dockerClient, err = client.NewEnvClient()
		if err != nil {
			panic(err)
		}
	})
}
