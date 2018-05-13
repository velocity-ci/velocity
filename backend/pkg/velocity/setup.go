package velocity

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Setup struct {
	BaseStep
	backupResolver BackupResolver
	repository     *GitRepository
	commitHash     string
}

func NewSetup() *Setup {
	return &Setup{
		BaseStep: BaseStep{
			Type:          "setup",
			OutputStreams: []string{"setup"},
		},
	}
}

func (s *Setup) Init(
	backupResolver BackupResolver,
	repository *GitRepository,
	commitHash string,
) {
	s.backupResolver = backupResolver
	s.repository = repository
	s.commitHash = commitHash
}

func (s *Setup) UnmarshalYamlInterface(y map[interface{}]interface{}) error {
	return nil
}

func (s Setup) GetDetails() string {
	return ""
}

func makeVelocityDirs() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	logrus.Infof("mkdir -p %s", fmt.Sprintf("%s/.velocityci/plugins", wd))

	os.MkdirAll(fmt.Sprintf("%s/.velocityci/plugins", wd), os.ModePerm)

	return nil
}

func (s *Setup) Execute(emitter Emitter, t *Task) error {

	t.RunID = fmt.Sprintf("vci-%s", time.Now().Format("060102150405"))

	writer := emitter.GetStreamWriter("setup")
	writer.SetStatus(StateRunning)

	// Clone repository if necessary
	if s.repository != nil {
		repo, err := Clone(s.repository, false, true, t.Git.Submodule, writer)
		if err != nil {
			logrus.Error(err)
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
			return err
		}
		logrus.Infof("Checking out %s", s.commitHash)
		repo.Checkout(s.commitHash)
		if err != nil {
			logrus.Error(err)
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
			return err
		}
		os.Chdir(repo.Directory)
	}

	if err := makeVelocityDirs(); err != nil {
		return err
	}

	// Resolve parameters
	parameters := map[string]Parameter{}
	for k, v := range getGitParams() {
		parameters[k] = v
		writer.Write([]byte(fmt.Sprintf("Set %s: %s", k, v.Value)))
	}

	// config
	for _, config := range t.Parameters {
		writer.Write([]byte(fmt.Sprintf("Resolving parameter %s", config.GetInfo())))
		params, err := config.GetParameters(writer, t, s.backupResolver)
		if err != nil {
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("could not resolve parameter: %v", err)))
			return fmt.Errorf("could not resolve %v", err)
		}
		for _, param := range params {
			parameters[param.Name] = param
			if param.IsSecret {
				writer.Write([]byte(fmt.Sprintf("Set %s: ***", param.Name)))
			} else {
				writer.Write([]byte(fmt.Sprintf("Set %s: %v", param.Name, param.Value)))
			}
		}
	}

	t.ResolvedParameters = parameters

	// Update params on steps
	for _, s := range t.Steps {
		s.SetParams(parameters)
	}

	// Login to docker registries
	authedRegistries := []DockerRegistry{}
	for _, registry := range t.Docker.Registries {
		r, err := dockerLogin(registry, writer, t.RunID, parameters, t.Docker.Registries)
		if err != nil || r.Address == "" {
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("could not login to Docker registry: %v", err)))
			return err
		}
		authedRegistries = append(authedRegistries, r)
		writer.Write([]byte(fmt.Sprintf("Authenticated with Docker registry: %s", r.Address)))
	}

	t.Docker.Registries = authedRegistries

	writer.SetStatus(StateSuccess)
	writer.Write([]byte("Setup success."))

	return nil
}

func (s *Setup) SetParams(params map[string]Parameter) error {
	return nil
}

func (s Setup) Validate(params map[string]Parameter) error {
	return nil
}

func getGitParams() map[string]Parameter {
	path, _ := os.Getwd()

	repo := &RawRepository{Directory: path}

	rawCommit := repo.GetCurrentCommitInfo()

	return map[string]Parameter{
		"GIT_COMMIT_LONG_SHA": {
			Value:    rawCommit.SHA,
			IsSecret: false,
		},
		"GIT_COMMIT_SHORT_SHA": {
			Value:    rawCommit.SHA[:7],
			IsSecret: false,
		},
		// "GIT_BRANCH": {
		// 	Value:    branch,
		// 	IsSecret: false,
		// },
		"GIT_DESCRIBE": {
			Value:    repo.GetDescribe(),
			IsSecret: false,
		},
		"GIT_COMMIT_AUTHOR": {
			Value:    rawCommit.AuthorEmail,
			IsSecret: false,
		},
		"GIT_COMMIT_MESSAGE": {
			Value:    rawCommit.Message,
			IsSecret: false,
		},
		"GIT_COMMIT_TIMESTAMP": {
			Value:    rawCommit.AuthorDate.String(),
			IsSecret: false,
		},
	}
}

func dockerLogin(registry DockerRegistry, writer io.Writer, RunID string, parameters map[string]Parameter, authConfigs []DockerRegistry) (r DockerRegistry, _ error) {

	type registryAuthConfig struct {
		Username      string `json:"username"`
		Password      string `json:"password"`
		ServerAddress string `json:"serverAddress"`
		Error         string `json:"error"`
		State         string `json:"state"`
	}

	bin, err := getBinary(registry.Use)
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

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), extraEnv...)

	cmdOutBytes, err := cmd.Output()
	if err != nil {
		return r, err
	}
	var dOutput registryAuthConfig
	json.Unmarshal(cmdOutBytes, &dOutput)

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
