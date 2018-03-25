package velocity

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
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
		repo, dir, err := GitClone(s.repository, false, true, t.Git.Submodule, writer)
		if err != nil {
			logrus.Error(err)
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
			return err
		}
		w, err := repo.Worktree()
		if err != nil {
			logrus.Error(err)
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
			return err
		}
		logrus.Infof("Checking out %s", s.commitHash)
		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(s.commitHash),
		})
		if err != nil {
			logrus.Error(err)
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
			return err
		}
		os.Chdir(dir)
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

	// We instance a new repository targeting the given path (the .git folder)
	r, err := git.PlainOpen(fmt.Sprintf("%s/", path))
	if err != nil {
		panic(err)
	}

	// ... retrieving the HEAD reference
	ref, err := r.Head()
	if err != nil {
		panic(err)
	}
	SHA := ref.Hash().String()
	shortSHA := SHA[:7]
	branch := ref.Name().Short()
	describe := shortSHA

	commit, err := r.CommitObject(ref.Hash())
	mParts := strings.Split(commit.Message, "-----END PGP SIGNATURE-----")
	message := mParts[0]
	if len(mParts) > 1 {
		message = mParts[1]
	}
	message = strings.TrimSpace(message)
	if err != nil {
		return map[string]Parameter{}
	}

	tags, _ := r.Tags()
	defer tags.Close()
	var lastTag *object.Tag
	for {
		t, err := tags.Next()
		if err == io.EOF {
			break
		}

		tObj, err := r.TagObject(t.Hash())
		if err != nil {
			panic(err)
		}

		c, _ := tObj.Commit()
		if c.Hash.String() == SHA {
			describe = tObj.Name
		}
		lastTag = tObj
	}

	if describe == shortSHA {
		if lastTag == nil {
			describe = shortSHA
		} else {
			describe = fmt.Sprintf("%s+%s", lastTag.Name, shortSHA)
		}
	}

	return map[string]Parameter{
		"GIT_COMMIT_LONG_SHA": {
			Value:    SHA,
			IsSecret: false,
		},
		"GIT_COMMIT_SHORT_SHA": {
			Value:    shortSHA,
			IsSecret: false,
		},
		"GIT_BRANCH": {
			Value:    branch,
			IsSecret: false,
		},
		"GIT_DESCRIBE": {
			Value:    describe,
			IsSecret: false,
		},
		"GIT_COMMIT_AUTHOR": {
			Value:    commit.Author.Email,
			IsSecret: false,
		},
		"GIT_COMMIT_MESSAGE": {
			Value:    message,
			IsSecret: false,
		},
		"GIT_COMMIT_TIMESTAMP": {
			Value:    commit.Committer.When.String(),
			IsSecret: false,
		},
		"GIT_COMMIT_TIMESTAMP_EPOCH": {
			Value:    strconv.FormatInt(commit.Committer.When.Unix(), 10),
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
