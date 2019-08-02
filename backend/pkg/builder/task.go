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
	eventBuildPrefix = fmt.Sprintf("%stask-", EventPrefix)
)

type ArchitectProject struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

type KnownHost struct {
	Entry string `json:"entry"`
}

type Task struct {
	*baseJob

	Project   ArchitectProject `json:"project"`
	KnownHost KnownHost        `json:"knownHost"`

	Task       *build.Task       `json:"task"`
	Branch     string            `json:"branch"`
	Commit     string            `json:"commit"`
	Parameters map[string]string `json:"parameters"`
}

func NewTask() *Task {
	return &Task{
		baseJob: &baseJob{
			Name: "task",
		},
	}
}

func (j *Task) Parse(payloadBytes []byte) error {
	return json.Unmarshal(payloadBytes, j)
}

func (j *Task) Do(ws *phoenix.Client) error {
	logging.GetLogger().Info("running task", zap.String("taskID", j.ID))

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
	logging.GetLogger().Info("completed task", zap.String("taskID", j.ID))
	os.Chdir(WorkspaceDir)
	return nil
}

func (j *Task) Stop(ws *phoenix.Client) error {

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
