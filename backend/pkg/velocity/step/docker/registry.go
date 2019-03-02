package docker

import (
	"strings"

	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

func GetAuthToken(image string, dockerRegistries []velocity.DockerRegistry) string {
	tagParts := strings.Split(image, "/")
	registry := tagParts[0]
	if strings.Contains(registry, ".") {
		// private
		for _, r := range dockerRegistries {
			if r.Address == registry {
				return r.AuthorizationToken
			}
		}
	} else {
		for _, r := range dockerRegistries {
			if strings.Contains(r.Address, "https://registry.hub.docker.com") || strings.Contains(r.Address, "https://index.docker.io") {
				return r.AuthorizationToken
			}
		}
	}

	return ""
}
