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
	eventBuildPrefix   = fmt.Sprintf("%sbuild-", EventPrefix)
	EventBuildNewEvent = fmt.Sprintf("%snew-event", eventBuildPrefix)
)

type Project struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

type Build struct {
	*baseJob

	ID      string  `json:"id"`
	Project Project `json:"project"`

	Task       *build.Task       `json:"buildTask"`
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
	Status     string    `json:"status"`
	Output     string    `json:"output"`
}

type BuildLogPayload struct {
	StreamID string          `json:"streamId"`
	Lines    []*BuildLogLine `json:"lines"`
}
