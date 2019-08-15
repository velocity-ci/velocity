// +build integration

package docker_test

import (
	"testing"
	"testing/iotest"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/build"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/docker"
)

func TestServiceRunner(t *testing.T) {
	writer := iotest.NewWriteLogger("serviceRunner", build.BlankWriter{})
	secrets := []string{}
	image := "busybox"
	config := &container.Config{
		Image:   image,
		Cmd:     []string{"ls"},
		Volumes: map[string]struct{}{},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{},
	}

	authConfigs := map[string]types.AuthConfig{}
	authTokens := map[string]string{}

	containerManager := docker.NewContainerManager("test", authConfigs, authTokens)

	containerManager.AddContainer(docker.NewContainer(
		writer,
		"test-run",
		image,
		nil,
		config,
		hostConfig,
		nil,
	))

	err := containerManager.Execute(secrets)
	assert.Nil(t, err)
}

func TestServiceRunnerInterrupt(t *testing.T) {
	writer := iotest.NewWriteLogger("serviceRunner", build.BlankWriter{})
	secrets := []string{}
	image := "busybox"
	config := &container.Config{
		Image:   image,
		Cmd:     []string{"sleep", "10000"},
		Volumes: map[string]struct{}{},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{},
	}

	authConfigs := map[string]types.AuthConfig{}
	authTokens := map[string]string{}

	containerManager := docker.NewContainerManager("test-interrupt", authConfigs, authTokens)

	containerManager.AddContainer(docker.NewContainer(
		writer,
		"test-run-interrupt",
		image,
		nil,
		config,
		hostConfig,
		nil,
	))

	go func() {
		for {
			time.Sleep(1 * time.Millisecond)
			if containerManager.IsRunning() {
				err := containerManager.Stop()
				assert.Nil(t, err)
				break
			}
		}
	}()

	err := containerManager.Execute(secrets)
	assert.Error(t, err)
}
