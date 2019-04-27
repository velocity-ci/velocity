package builder

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"

	"github.com/velocity-ci/velocity/backend/pkg/git"

	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/build"
	"go.uber.org/zap"
)

const WorkspaceDir = "/opt/velocityci"

// Builder -> Architect (Build events)
var (
	eventBuildPrefix = fmt.Sprintf("%sbuild-", EventPrefix)
)

type ArchitectProject struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

type KnownHost struct {
	Entry string `json:"entry"`
}

type Build struct {
	*baseJob

	Project   ArchitectProject `json:"project"`
	KnownHost KnownHost        `json:"knownHost"`

	Task       *build.Task       `json:"task"`
	Branch     string            `json:"branch"`
	Commit     string            `json:"commit"`
	Parameters map[string]string `json:"parameters"`
}

func NewBuild() *Build {
	return &Build{
		baseJob: &baseJob{
			Name: "build",
		},
	}
}

func (j *Build) Parse(payloadBytes []byte) error {
	return json.Unmarshal(payloadBytes, j)
}

func (j *Build) Do(ws *phoenix.Client) error {
	logging.GetLogger().Info("running build", zap.String("buildID", j.ID))

	emitter := NewEmitter(ws)
	backupResolver := NewParameterResolver(j.Parameters)

	j.Task.UpdateSetup(backupResolver, &git.Repository{
		Address:    j.Project.Address,
		PrivateKey: j.Project.PrivateKey,
	}, j.Branch, j.Commit)

	// TODO: add knownhost file management

	j.Task.Execute(emitter)

	wd, _ := os.Getwd()
	if strings.HasPrefix(wd, WorkspaceDir) {
		os.RemoveAll(wd)
	}
	logging.GetLogger().Info("completed build", zap.String("buildID", j.ID))
	os.Chdir(WorkspaceDir)
	return nil
}

func (j *Build) Stop(ws *phoenix.Client) error {

	return nil
}

type BuildLogLine struct {
	Timestamp  time.Time `json:"timestamp"`
	LineNumber int       `json:"lineNumber"`
	// Status     string    `json:"status"`
	Output string `json:"output"`
}

type TaskUpdatePayload struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

type StepUpdatePayload struct {
	ID    string `json:"id"`
	State string `json:"state"`
}

type StreamNewLogLinePayload struct {
	ID string `json:"id"`
	// Lines []*BuildLogLine `json:"lines"`
	Lines []interface{} `json:"lines"`
}
