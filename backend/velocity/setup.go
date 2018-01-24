package velocity

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type Setup struct {
	BaseStep
	task           *Task
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

func (s *Setup) UnmarshalYamlInterface(y map[interface{}]interface{}) error {
	return nil
}

func (s Setup) GetDetails() string {
	return ""
}

func (s *Setup) Init(
	task *Task,
	backupResolver BackupResolver,
	repository *GitRepository,
	commitHash string,
) {
	s.task = task
	s.backupResolver = backupResolver
	s.repository = repository
	s.commitHash = commitHash
}

func (s *Setup) Execute(emitter Emitter, params map[string]Parameter) error {
	s.task.runID = fmt.Sprintf("vci-%s", time.Now().Format("060102150405"))

	writer := emitter.GetStreamWriter("setup")
	writer.SetStatus(StateRunning)

	// Clone repository if necessary
	if s.repository != nil {
		repo, dir, err := GitClone(s.repository, false, true, s.task.Git.Submodule, writer)
		if err != nil {
			log.Println(err)
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
			return err
		}
		w, err := repo.Worktree()
		if err != nil {
			log.Println(err)
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
			return err
		}
		log.Printf("Checking out %s", s.commitHash)
		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(s.commitHash),
		})
		if err != nil {
			log.Println(err)
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
			return err
		}
		os.Chdir(dir)
	}

	// Resolve parameters
	parameters := map[string]Parameter{}
	for k, v := range getGitParams() {
		parameters[k] = v
		writer.Write([]byte(fmt.Sprintf("Set %s: %s", k, v.Value)))
	}

	// config
	for _, config := range s.task.Parameters {
		writer.Write([]byte(fmt.Sprintf("Resolving parameter %s", config.GetInfo())))
		params, err := config.GetParameters(writer, s.task.runID, s.backupResolver)
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

	// Update params on steps
	for _, s := range s.task.Steps {
		s.SetParams(parameters)
	}

	// Login to docker registries
	authedRegistries := []DockerRegistry{}
	for _, registry := range s.task.Docker.Registries {
		r, err := dockerLogin(registry, writer, s.task.runID, parameters)
		if err != nil {
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("could not login to Docker registry: %v", err)))
			return err
		}
		authedRegistries = append(authedRegistries, r)
		writer.Write([]byte(fmt.Sprintf("Authenticated with Docker registry: %s (%s)", r.Address, r.AuthorizationToken)))
	}

	for _, s := range s.task.Steps {
		s.SetDockerRegistries(authedRegistries)
	}

	log.Printf("%+v", s.task.Docker.Registries)

	writer.SetStatus(StateSuccess)
	writer.Write([]byte(""))
	writer.Write([]byte("Setup success.\n"))

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
		"GIT_COMMIT_LONG_SHA": Parameter{
			Value:    SHA,
			IsSecret: false,
		},
		"GIT_COMMIT_SHORT_SHA": Parameter{
			Value:    shortSHA,
			IsSecret: false,
		},
		"GIT_BRANCH": Parameter{
			Value:    branch,
			IsSecret: false,
		},
		"GIT_DESCRIBE": Parameter{
			Value:    describe,
			IsSecret: false,
		},
		"GIT_COMMIT_AUTHOR": Parameter{
			Value:    commit.Author.Email,
			IsSecret: false,
		},
		"GIT_COMMIT_MESSAGE": Parameter{
			Value:    message,
			IsSecret: false,
		},
		"GIT_COMMIT_TIMESTAMP": Parameter{
			Value:    commit.Committer.When.String(),
			IsSecret: false,
		},
		"GIT_COMMIT_TIMESTAMP_EPOCH": Parameter{
			Value:    strconv.FormatInt(commit.Committer.When.Unix(), 10),
			IsSecret: false,
		},
	}
}

func dockerLogin(registry DockerRegistry, writer io.Writer, runID string, parameters map[string]Parameter) (DockerRegistry, error) {
	env := []string{
		fmt.Sprintf("address=%s", registry.Address),
	}
	for k, v := range registry.Arguments {
		for _, pV := range parameters {
			v = strings.Replace(v, fmt.Sprintf("${%s}", pV.Name), pV.Value, -1)
			k = strings.Replace(k, fmt.Sprintf("${%s}", pV.Name), pV.Value, -1)
		}
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	containerConfig := &container.Config{
		Image: registry.Use,
		Env:   env,
		Volumes: map[string]struct{}{
			"/var/run/docker.sock": struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			"/var/run/docker.sock:/var/run/docker.sock",
		},
	}

	var wg sync.WaitGroup
	cli, _ := client.NewEnvClient()
	ctx := context.Background()

	networkResp, err := cli.NetworkCreate(ctx, runID, types.NetworkCreate{
		Labels: map[string]string{"owner": "velocity-ci"},
	})
	if err != nil {
		log.Println(err)
	}
	blankEmitter := NewBlankEmitter()
	w := blankEmitter.GetStreamWriter("setup")
	sR := newServiceRunner(
		cli,
		ctx,
		w,
		&wg,
		map[string]Parameter{},
		fmt.Sprintf("%s-%s", runID, "dLogin"),
		registry.Use,
		nil,
		containerConfig,
		hostConfig,
		nil,
		networkResp.ID,
	)

	sR.PullOrBuild()
	sR.Create()
	stopServicesChannel := make(chan string, 32)
	wg.Add(1)
	go sR.Run(stopServicesChannel)
	_ = <-stopServicesChannel

	logsResp, err := sR.dockerCli.ContainerLogs(
		sR.context,
		sR.containerID,
		types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: false},
	)
	if err != nil {
		log.Printf("param container %s logs err: %s", sR.containerID, err)
	}
	headerBytes := make([]byte, 8)
	logsResp.Read(headerBytes)
	content, _ := ioutil.ReadAll(logsResp)
	logsResp.Close()
	sR.Stop()
	wg.Wait()
	err = cli.NetworkRemove(ctx, networkResp.ID)
	if err != nil {
		log.Printf("network %s remove err: %s", networkResp.ID, err)
	}

	var dOutput dockerLoginOutput
	json.Unmarshal(content, &dOutput)
	if dOutput.State != "success" {
		return registry, fmt.Errorf(dOutput.Error)
	}

	registry.AuthorizationToken = dOutput.AuthorizationToken
	return registry, nil
}

type dockerLoginOutput struct {
	State              string `json:"state"`
	Error              string `json:"error"`
	AuthorizationToken string `json:"authToken"`
}