package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gosimple/slug"
	"github.com/velocity-ci/velocity/backend/pkg/auth"
	"github.com/velocity-ci/velocity/backend/pkg/git"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/docker"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"go.uber.org/zap"
)

type Setup struct {
	BaseStep
	backupResolver BackupResolver
	repository     *git.Repository
	branch         string
	commitHash     string
}

func getUniqueWorkspace(r *git.Repository) (string, error) {
	dir := fmt.Sprintf("%s/_%s-%s",
		"/tmp",
		slug.Make(r.Address),
		auth.RandomString(8),
	)

	err := os.RemoveAll(dir)
	if err != nil {
		logging.GetLogger().Fatal("could not create unique workspace", zap.Error(err))
		return "", err
	}
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		logging.GetLogger().Fatal("could not create unique workspace", zap.Error(err))
		return "", err
	}

	return dir, nil
}

func NewStepSetup(
	resolver BackupResolver,
	repository *git.Repository,
	branch,
	commitSha string,
) *Setup {
	return &Setup{
		BaseStep:       newBaseStep("setup", []string{"setup"}),
		backupResolver: resolver,
		repository:     repository,
		branch:         branch,
		commitHash:     commitSha,
	}
}

func (s Setup) GetDetails() string {
	return ""
}

func makeVelocityDirs(projectRoot string) error {
	os.MkdirAll(filepath.Join(projectRoot, ".velocityci/plugins"), os.ModePerm)

	return nil
}

func (s *Setup) Execute(emitter Emitter, t *Task) error {
	// t.ID = fmt.Sprintf("vci-%s", uuid.NewV4().String())
	// logging.GetLogger().Debug("set run id", zap.String("runID", t.ID))

	writer, err := s.GetStreamWriter(emitter, "setup")
	if err != nil {
		return err
	}
	defer writer.Close()
	writer.SetStatus(StateRunning)

	// Clone repository if necessary
	if s.repository != nil {
		dir, _ := getUniqueWorkspace(s.repository)
		repo, err := git.Clone(
			s.repository,
			&git.CloneOptions{Bare: false, Submodule: true, Commit: s.commitHash},
			dir,
			writer,
		)
		if err != nil {
			logging.GetLogger().Error("could not clone repository", zap.Error(err))
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, docker.ColorFmt(docker.ANSIError, "-> failed: %s"), err)
			return err
		}
		err = repo.Checkout(s.commitHash)
		if err != nil {
			logging.GetLogger().Error("could not checkout", zap.Error(err), zap.String("commit", s.commitHash))
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, docker.ColorFmt(docker.ANSIError, "-> failed: %s"), err)
			return err
		}
		os.Chdir(repo.Directory)
		t.ProjectRoot = repo.Directory
	}

	if err := makeVelocityDirs(t.ProjectRoot); err != nil {
		return err
	}

	// Resolve parameters
	t.parameters = map[string]*Parameter{}
	basicParams, err := GetGlobalParams(writer, t.ProjectRoot, s.branch)
	if err != nil {
		return err
	}
	for k, v := range basicParams {
		t.parameters[k] = &v
		writer.Write([]byte(fmt.Sprintf("Set %s: %s", k, v.Value)))
	}
	for _, configParam := range t.Config.Parameters {
		resolvedParams, err := resolveConfigParameter(configParam, s.backupResolver, t.ProjectRoot, writer)
		if err != nil {
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, docker.ColorFmt(docker.ANSIError, "-> could not resolve parameter: %s"), err)
			return fmt.Errorf("could not resolve %v", err)
		}
		for _, param := range resolvedParams {
			t.parameters[param.Name] = param
			if param.IsSecret {
				fmt.Fprintf(writer, docker.ColorFmt(docker.ANSIInfo, "-> set %s: ***"), param.Name)
			} else {
				fmt.Fprintf(writer, docker.ColorFmt(docker.ANSIInfo, "-> set %s: %v"), param.Name, param.Value)
			}
		}
	}

	// Login to docker registries
	authedRegistries := []DockerRegistry{}
	for _, registry := range t.Docker.Registries {
		r, err := dockerLogin(registry, writer, t)
		if err != nil || r.Address == "" {
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, docker.ColorFmt(docker.ANSIError, "-> could not login to Docker registry: %s"), err)

			return err
		}
		authedRegistries = append(authedRegistries, r)
		fmt.Fprintf(writer, docker.ColorFmt(docker.ANSIInfo, "-> authenticated with Docker registry: %s"), r.Address)
	}

	t.Docker.Registries = authedRegistries

	writer.SetStatus(StateSuccess)
	writer.Write([]byte("Setup success."))

	return nil
}

func (s *Setup) SetParams(params map[string]*Parameter) error {
	return nil
}

func (s Setup) Validate(params map[string]Parameter) error {
	return nil
}

func GetGlobalParams(writer io.Writer, projectRoot, branch string) (map[string]Parameter, error) {
	params := map[string]Parameter{}

	repo := &git.RawRepository{Directory: projectRoot}

	if repo.IsDirty() {
		fmt.Fprintf(writer, docker.ColorFmt(docker.ANSIWarn, "-> Project files are dirty. Build repeatability is not guaranteed."))
	}

	rawCommit, _ := repo.GetCurrentCommitInfo()

	buildTimestamp := time.Now().UTC()
	params["GIT_BRANCH"] = Parameter{
		Value:    branch,
		IsSecret: false,
	}
	params["GIT_COMMIT_LONG_SHA"] = Parameter{
		Value:    rawCommit.SHA,
		IsSecret: false,
	}
	params["GIT_COMMIT_SHORT_SHA"] = Parameter{
		Value:    rawCommit.SHA[:7],
		IsSecret: false,
	}
	params["GIT_DESCRIBE"] = Parameter{
		Value:    repo.GetDescribe(),
		IsSecret: false,
	}
	params["GIT_DESCRIBE_ALL"] = Parameter{
		Value:    repo.GetDescribeAll(),
		IsSecret: false,
	}
	params["GIT_COMMIT_AUTHOR"] = Parameter{
		Value:    rawCommit.AuthorEmail,
		IsSecret: false,
	}
	params["GIT_COMMIT_MESSAGE"] = Parameter{
		Value:    rawCommit.Message,
		IsSecret: false,
	}
	params["GIT_COMMIT_TIMESTAMP_RFC3339"] = Parameter{
		Value:    rawCommit.AuthorDate.UTC().Format(time.RFC3339),
		IsSecret: false,
	}
	params["GIT_COMMIT_TIMESTAMP_RFC822"] = Parameter{
		Value:    rawCommit.AuthorDate.UTC().Format(time.RFC822),
		IsSecret: false,
	}
	params["BUILD_TIMESTAMP_RFC3339"] = Parameter{
		Value:    buildTimestamp.Format(time.RFC3339),
		IsSecret: false,
	}
	params["BUILD_TIMESTAMP_RFC822"] = Parameter{
		Value:    buildTimestamp.Format(time.RFC822),
		IsSecret: false,
	}

	return params, nil
}
