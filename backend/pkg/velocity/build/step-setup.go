package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/auth"
	"github.com/velocity-ci/velocity/backend/pkg/git"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
	"go.uber.org/zap"
)

type Setup struct {
	BaseStep
	backupResolver BackupResolver
	repository     *git.Repository
	commitHash     string
}

func getUniqueWorkspace(r *git.Repository) (string, error) {
	dir := fmt.Sprintf("%s/_%s-%s",
		"",
		// slug.Make(r.Address),
		auth.RandomString(8),
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
	commitSha string,
) *Setup {
	return &Setup{
		BaseStep:       newBaseStep("setup", []string{"setup"}),
		backupResolver: resolver,
		repository:     repository,
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

func (s *Setup) Execute(emitter out.Emitter, t *Task) error {
	t.RunID = fmt.Sprintf("vci-%s", uuid.NewV4().String())
	logging.GetLogger().Debug("set run id", zap.String("runID", t.RunID))

	writer := emitter.GetStreamWriter("setup")
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
			fmt.Fprintf(writer, out.ColorFmt(out.ANSIError, "-> failed: %s"), err)
			return err
		}
		err = repo.Checkout(s.commitHash)
		if err != nil {
			logging.GetLogger().Error("could not checkout", zap.Error(err), zap.String("commit", s.commitHash))
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, out.ColorFmt(out.ANSIError, "-> failed: %s"), err)
			return err
		}
		os.Chdir(repo.Directory)
		t.ProjectRoot = repo.Directory
	}

	if err := makeVelocityDirs(t.ProjectRoot); err != nil {
		return err
	}

	// Resolve parameters
	basicParams, err := GetGlobalParams(writer, t.ProjectRoot)
	if err != nil {
		return err
	}
	for k, v := range basicParams {
		t.Parameters[k] = &v
		writer.Write([]byte(fmt.Sprintf("Set %s: %s", k, v.Value)))
	}
	for configParam := range t.Config.Parameters {
		resolvedParams, err := resolveConfigParameter(configParam, s.backupResolver, t.ProjectRoot, writer)
		if err != nil {
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, out.ColorFmt(out.ANSIError, "-> could not resolve parameter: %s"), err)
			return fmt.Errorf("could not resolve %v", err)
		}
		for _, param := range resolvedParams {
			t.Parameters[param.Name] = param
			if param.IsSecret {
				fmt.Fprintf(writer, out.ColorFmt(out.ANSIInfo, "-> set %s: ***"), param.Name)
			} else {
				fmt.Fprintf(writer, out.ColorFmt(out.ANSIInfo, "-> set %s: %v"), param.Name, param.Value)
			}
		}
	}

	// config
	// for _, config := range t.Parameters {
	// 	writer.Write([]byte(fmt.Sprintf("-> resolving parameter %s", config.GetInfo())))
	// 	params, err := config.GetParameters(writer, t, s.backupResolver)
	// 	if err != nil {
	// 		writer.SetStatus(StateFailed)
	// 		fmt.Fprintf(writer, out.ColorFmt(out.ANSIError, "-> could not resolve parameter: %s"), err)
	// 		return fmt.Errorf("could not resolve %v", err)
	// 	}
	// 	for _, param := range params {
	// 		parameters[param.Name] = param
	// 		if param.IsSecret {
	// 			fmt.Fprintf(writer, out.ColorFmt(out.ANSIInfo, "-> set %s: ***"), param.Name)
	// 		} else {
	// 			fmt.Fprintf(writer, out.ColorFmt(out.ANSIInfo, "-> set %s: %v"), param.Name, param.Value)
	// 		}
	// 	}
	// }

	// t.ResolvedParameters = parameters

	// // Update params on steps
	// for _, s := range t.Steps {
	// 	s.SetParams(parameters)
	// }

	// Login to docker registries
	authedRegistries := []DockerRegistry{}
	for _, registry := range t.Docker.Registries {
		r, err := dockerLogin(registry, writer, t)
		if err != nil || r.Address == "" {
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, out.ColorFmt(out.ANSIError, "-> could not login to Docker registry: %s"), err)

			return err
		}
		authedRegistries = append(authedRegistries, r)
		fmt.Fprintf(writer, out.ColorFmt(out.ANSIInfo, "-> authenticated with Docker registry: %s"), r.Address)
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

func GetGlobalParams(writer io.Writer, projectRoot string) (map[string]Parameter, error) {
	params := map[string]Parameter{}
	// cwd, err := os.Getwd()
	// if err != nil {
	// 	return params, err
	// }
	// projectRoot, err := findProjectRoot(cwd, []string{})
	// if err != nil {
	// 	return params, err
	// }

	repo := &git.RawRepository{Directory: projectRoot}

	if repo.IsDirty() {
		fmt.Fprintf(writer, out.ColorFmt(out.ANSIWarn, "-> Project files are dirty. Build repeatability is not guaranteed."))
	}

	rawCommit, _ := repo.GetCurrentCommitInfo()

	buildTimestamp := time.Now().UTC()

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