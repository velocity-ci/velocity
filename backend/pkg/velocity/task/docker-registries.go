package task

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/velocity-ci/velocity/backend/pkg/exec"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
	"go.uber.org/zap"
)

type TaskDocker struct {
	Registries []DockerRegistry `json:"registries"`
}

type DockerRegistry struct {
	Address            string            `json:"address"`
	Use                string            `json:"use"`
	Arguments          map[string]string `json:"arguments"`
	AuthorizationToken string            `json:"authToken"`
}

func dockerLogin(registry DockerRegistry, writer io.Writer, task *Task, parameters map[string]Parameter) (r DockerRegistry, _ error) {

	type registryAuthConfig struct {
		Username      string `json:"username"`
		Password      string `json:"password"`
		ServerAddress string `json:"serverAddress"`
		Error         string `json:"error"`
		State         string `json:"state"`
	}

	bin, err := getBinary(task.ProjectRoot, registry.Use, writer)
	if err != nil {
		return r, err
	}

	extraEnv := []string{}
	for k, v := range registry.Arguments {
		for _, pV := range parameters {
			v = strings.Replace(v, fmt.Sprintf("${%s}", pV.Name), pV.Value, -1)
			k = strings.Replace(k, fmt.Sprintf("${%s}", pV.Name), pV.Value, -1)
		}
		extraEnv = append(extraEnv, fmt.Sprintf("%s=%s", k, v))
	}

	s := exec.Run([]string{bin}, "", append(os.Environ(), extraEnv...), out.BlankWriter{})
	if s.Error != nil {
		return r, err
	}

	var dOutput registryAuthConfig
	json.Unmarshal([]byte(s.Stdout[0]), &dOutput)

	if dOutput.State != "success" {
		return r, fmt.Errorf("registry auth error: %s", dOutput.Error)
	}

	cli, _ := client.NewEnvClient()
	ctx := context.Background()
	_, err = cli.RegistryLogin(ctx, types.AuthConfig{
		Username:      dOutput.Username,
		Password:      dOutput.Password,
		ServerAddress: dOutput.ServerAddress,
	})
	if err != nil {
		return r, err
	}

	authConfig := types.AuthConfig{
		Username: dOutput.Username,
		Password: dOutput.Password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return r, err
	}
	registry.AuthorizationToken = base64.URLEncoding.EncodeToString(encodedJSON)
	registry.Address = dOutput.ServerAddress

	return registry, nil
}

type dockerLoginOutput struct {
	State              string `json:"state"`
	Error              string `json:"error"`
	AuthorizationToken string `json:"authToken"`
	Address            string `json:"address"`
}

func GetAuthToken(image string, dockerRegistries []DockerRegistry) string {
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

func GetAuthConfigsMap(dockerRegistries []DockerRegistry) map[string]types.AuthConfig {
	authConfigs := map[string]types.AuthConfig{}
	for _, r := range dockerRegistries {
		jsonAuthConfig, err := base64.URLEncoding.DecodeString(r.AuthorizationToken)
		if err != nil {
			logging.GetLogger().Error(
				"could not decode registry auth config",
				zap.String("err", err.Error()),
				zap.String("registry", r.Address),
			)

		}
		var authConfig types.AuthConfig
		err = json.Unmarshal(jsonAuthConfig, &authConfig)
		authConfigs[r.Address] = authConfig
	}

	return authConfigs
}
