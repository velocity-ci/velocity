package docker

import "os"

func init() {
	// Minimum supported version as recommended by https://docs.docker.com/develop/sdk/#api-version-matrix
	os.Setenv("DOCKER_API_VERSION", "1.24")
}
