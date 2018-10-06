package velocity

import (
	"fmt"
	"os"
	"time"

	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
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

	os.MkdirAll(fmt.Sprintf("%s/.velocityci/plugins", wd), os.ModePerm)

	return nil
}

func (s *Setup) Execute(emitter Emitter, t *Task) error {
	t.RunID = fmt.Sprintf("vci-%s", uuid.NewV4().String())
	GetLogger().Debug("set run id", zap.String("runID", t.RunID))

	writer := emitter.GetStreamWriter("setup")
	defer writer.Close()
	writer.SetStatus(StateRunning)

	// Clone repository if necessary
	if s.repository != nil {
		repo, err := Clone(s.repository, writer, &CloneOptions{Bare: false, Full: true, Submodule: true, Commit: s.commitHash})
		if err != nil {
			GetLogger().Error("could not clone repository", zap.Error(err))
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, colorFmt(ansiError, "-> failed: %s"), err)
			return err
		}
		err = repo.Checkout(s.commitHash)
		if err != nil {
			GetLogger().Error("could not checkout", zap.Error(err), zap.String("commit", s.commitHash))
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, colorFmt(ansiError, "-> failed: %s"), err)
			return err
		}
		os.Chdir(repo.Directory)
		t.ProjectRoot = repo.Directory
	}

	if err := makeVelocityDirs(); err != nil {
		return err
	}

	// Resolve parameters
	parameters := map[string]Parameter{}
	for k, v := range GetBasicParams() {
		parameters[k] = v
		writer.Write([]byte(fmt.Sprintf("Set %s: %s", k, v.Value)))
	}

	// config
	for _, config := range t.Parameters {
		writer.Write([]byte(fmt.Sprintf("-> resolving parameter %s", config.GetInfo())))
		params, err := config.GetParameters(writer, t, s.backupResolver)
		if err != nil {
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, colorFmt(ansiError, "-> could not resolve parameter: %s"), err)
			return fmt.Errorf("could not resolve %v", err)
		}
		for _, param := range params {
			parameters[param.Name] = param
			if param.IsSecret {
				fmt.Fprintf(writer, colorFmt(ansiInfo, "-> set %s: ***"), param.Name)
			} else {
				fmt.Fprintf(writer, colorFmt(ansiInfo, "-> set %s: %v"), param.Name, param.Value)
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
		r, err := dockerLogin(registry, writer, t, parameters)
		if err != nil || r.Address == "" {
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, colorFmt(ansiError, "-> could not login to Docker registry: %s"), err)

			return err
		}
		authedRegistries = append(authedRegistries, r)
		fmt.Fprintf(writer, colorFmt(ansiInfo, "-> authenticated with Docker registry: %s"), r.Address)
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

func GetBasicParams() map[string]Parameter {
	cwd, err := os.Getwd()
	if err != nil {
		// return
		GetLogger().Error("could not get cwd", zap.Error(err))
	}
	projectRoot, err := findProjectRoot(cwd)
	if err != nil {
		// return
		GetLogger().Error("could not find project root", zap.Error(err))
	}

	repo := &RawRepository{Directory: projectRoot}

	if repo.IsDirty() {
		// pass in writer to output dirty warning
	}

	rawCommit, _ := repo.GetCurrentCommitInfo()

	buildTimestamp := time.Now().UTC()

	return map[string]Parameter{
		"GIT_COMMIT_LONG_SHA": {
			Value:    rawCommit.SHA,
			IsSecret: false,
		},
		"GIT_COMMIT_SHORT_SHA": {
			Value:    rawCommit.SHA[:7],
			IsSecret: false,
		},
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
		"GIT_COMMIT_TIMESTAMP_RFC3339": {
			Value:    rawCommit.AuthorDate.UTC().Format(time.RFC3339),
			IsSecret: false,
		},
		"GIT_COMMIT_TIMESTAMP_RFC822": {
			Value:    rawCommit.AuthorDate.UTC().Format(time.RFC822),
			IsSecret: false,
		},
		"BUILD_TIMESTAMP_RFC3339": {
			Value:    buildTimestamp.Format(time.RFC3339),
			IsSecret: false,
		},
		"BUILD_TIMESTAMP_RFC822": {
			Value:    buildTimestamp.Format(time.RFC822),
			IsSecret: false,
		},
	}
}
