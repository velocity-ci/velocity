package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/gosimple/slug"
	"github.com/velocity-ci/velocity/backend/pkg/auth"
	"github.com/velocity-ci/velocity/backend/pkg/git"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
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
	return "n/a"
}

func makeVelocityDirs(projectRoot string) error {
	os.MkdirAll(filepath.Join(projectRoot, ".velocityci/plugins"), os.ModePerm)

	return nil
}

func sortedParameterKeys(params map[string]Parameter) []string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}

func (s *Setup) Execute(emitter Emitter, t *Task) error {
	// t.ID = fmt.Sprintf("vci-%s", uuid.NewV4().String())
	// logging.GetLogger().Debug("set run id", zap.String("runID", t.ID))

	writer, err := s.GetStreamWriter(emitter, "setup")
	if err != nil {
		return err
	}
	defer writer.Close()
	writer.SetStatus(StateBuilding)
	fmt.Fprintf(writer, "\r")

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
			fmt.Fprintf(writer, output.ColorFmt(output.ANSIError, "-> Could not clone repository: %s", "\n"), err)
			return err
		}
		err = repo.Checkout(s.commitHash)
		if err != nil {
			logging.GetLogger().Error("could not checkout", zap.Error(err), zap.String("commit", s.commitHash))
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, output.ColorFmt(output.ANSIError, "Could not checkout ref: %s", "\n"), err)
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
	tabWriter := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)
	for _, k := range sortedParameterKeys(basicParams) {
		v := basicParams[k]
		t.parameters[k] = &v
		fmt.Fprintf(tabWriter, "Set %s\t%s\n", k, v.Value)
	}
	tabWriter.Flush()
	for _, configParam := range t.Blueprint.Parameters {
		resolvedParams, err := resolveConfigParameter(configParam, s.backupResolver, t.ProjectRoot, writer)
		if err != nil {
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, output.ColorFmt(output.ANSIError, "Could not resolve parameter: %s", "\n"), err)
			return fmt.Errorf("could not resolve %v", err)
		}
		for _, param := range resolvedParams {
			t.parameters[param.Name] = param
			if param.IsSecret {
				fmt.Fprintf(writer, "Set %s: ***\n", param.Name)
			} else {
				fmt.Fprintf(writer, "Set %s: %v\n", param.Name, param.Value)
			}
		}
	}

	// Login to docker registries
	authedRegistries := []DockerRegistry{}
	for _, registry := range t.Docker.Registries {
		r, err := dockerLogin(registry, writer, t)
		if err != nil || r.Address == "" {
			writer.SetStatus(StateFailed)
			fmt.Fprintf(writer, output.ColorFmt(output.ANSIError, "Could not login to Docker registry: %s", "\n"), err)

			return err
		}
		authedRegistries = append(authedRegistries, r)
		fmt.Fprintf(writer, output.ColorFmt(output.ANSIInfo, "Authenticated with Docker registry: %s", "\n"), r.Address)
	}

	t.Docker.Registries = authedRegistries

	writer.SetStatus(StateSuccess)
	writer.Write([]byte("Setup success.\n"))

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
		fmt.Fprintf(writer, output.ColorFmt(output.ANSIWarn, "Project files are dirty. Build repeatability is not guaranteed.", "\n"))
	}

	rawCommit, err := repo.GetCurrentCommitInfo()
	if err != nil {
		return params, err
	}
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return params, err
	}

	buildTimestamp := time.Now().UTC()
	params["git.branch"] = Parameter{
		Value:    branch,
		IsSecret: false,
	}
	params["git.commit.sha.long"] = Parameter{
		Value:    rawCommit.SHA,
		IsSecret: false,
	}
	params["git.commit.sha.short"] = Parameter{
		Value:    rawCommit.SHA[:7],
		IsSecret: false,
	}
	params["git.describe"] = Parameter{
		Value:    repo.GetDescribe(),
		IsSecret: false,
	}
	params["git.describe.all"] = Parameter{
		Value:    repo.GetDescribeAll(),
		IsSecret: false,
	}
	params["git.commit.author"] = Parameter{
		Value:    rawCommit.AuthorEmail,
		IsSecret: false,
	}
	params["git.commit.message"] = Parameter{
		Value:    rawCommit.Message,
		IsSecret: false,
	}
	params["git.commit.rfc3339"] = Parameter{
		Value:    rawCommit.AuthorDate.UTC().Format(time.RFC3339),
		IsSecret: false,
	}
	params["git.commit.rfc822"] = Parameter{
		Value:    rawCommit.AuthorDate.UTC().Format(time.RFC822),
		IsSecret: false,
	}
	params["git.commit.rfc3339.clean"] = Parameter{
		Value:    reg.ReplaceAllString(rawCommit.AuthorDate.UTC().Format(time.RFC3339), ""),
		IsSecret: false,
	}
	params["git.commit.rfc822.clean"] = Parameter{
		Value:    reg.ReplaceAllString(rawCommit.AuthorDate.UTC().Format(time.RFC822), ""),
		IsSecret: false,
	}
	params["build.rfc3339"] = Parameter{
		Value:    buildTimestamp.Format(time.RFC3339),
		IsSecret: false,
	}
	params["build.rfc3339.clean"] = Parameter{
		Value:    reg.ReplaceAllString(buildTimestamp.Format(time.RFC3339), ""),
		IsSecret: false,
	}
	params["build.rfc822"] = Parameter{
		Value:    buildTimestamp.Format(time.RFC822),
		IsSecret: false,
	}
	params["build.rfc822.clean"] = Parameter{
		Value:    reg.ReplaceAllString(buildTimestamp.Format(time.RFC822), ""),
		IsSecret: false,
	}

	return params, nil
}
