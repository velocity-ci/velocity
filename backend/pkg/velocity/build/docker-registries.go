package build

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

func dockerLogin(registry DockerRegistry, writer io.Writer, task *Task) (r DockerRegistry, _ error) {

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
		for _, pV := range task.parameters {
			v = strings.Replace(v, fmt.Sprintf("${%s}", pV.Name), pV.Value, -1)
			k = strings.Replace(k, fmt.Sprintf("${%s}", pV.Name), pV.Value, -1)
		}
		extraEnv = append(extraEnv, fmt.Sprintf("%s=%s", k, v))
	}

	s := exec.Run([]string{bin}, "", append(os.Environ(), extraEnv...), BlankWriter{})
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

func GetAddressAuthTokensMap(dockerRegistries []DockerRegistry) (r map[string]string) {
	r = map[string]string{}
	for _, dR := range dockerRegistries {
		r[dR.Address] = dR.AuthorizationToken
	}
	return r
}
