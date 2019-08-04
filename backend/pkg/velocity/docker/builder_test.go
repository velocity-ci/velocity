package docker_test

import (
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"

	"github.com/docker/docker/api/types"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/build"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/docker"
)

func TestImageBuilderBuild(t *testing.T) {
	builder := docker.NewImageBuilder()

	writer := iotest.NewWriteLogger("builder", build.BlankWriter{})
	secrets := []string{}
	buildContext := "test/"
	dockerfile := "Dockerfile"
	tags := []string{}
	authConfigs := map[string]types.AuthConfig{}

	err := builder.Build(writer, secrets, buildContext, dockerfile, tags, authConfigs)
	assert.Nil(t, err)
}

func TestImageBuilderBuildInterrupt(t *testing.T) {
	builder := docker.NewImageBuilder()

	writer := iotest.NewWriteLogger("builder", build.BlankWriter{})
	secrets := []string{}
	buildContext := "test/"
	dockerfile := "long.Dockerfile"
	tags := []string{}
	authConfigs := map[string]types.AuthConfig{}

	go func() {
		for {
			if builder.IsRunning() {
				err := builder.Stop()
				assert.Nil(t, err)
				break
			}
		}
	}()

	err := builder.Build(writer, secrets, buildContext, dockerfile, tags, authConfigs)
	assert.Error(t, err)
}
