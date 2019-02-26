package builder

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

// Builder -> Architect (Build events)
var (
	eventBuildPrefix   = fmt.Sprintf("%sbuild-", EventPrefix)
	EventBuildNewEvent = fmt.Sprintf("%snew-event", eventBuildPrefix)
)

type Build struct {
	*baseJob
	Build   *build.Build    `json:"build"`
	Steps   []*build.Step   `json:"steps"`
	Streams []*build.Stream `json:"streams"`
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
	velocity.GetLogger().Info("running build", zap.String("buildID", j.Build.ID))

	emitter := NewEmitter(ws, j.Build)

	backupResolver := NewParameterResolver(j.Build.Parameters)

	vT := j.Build.Task.VTask

	for i, step := range vT.Steps {
		bStep := j.Steps[i]
		velocity.GetLogger().Debug("running step", zap.String("stepID", bStep.ID))
		emitter.SetStepAndStreams(bStep, j.Streams)

		if step.GetType() == "setup" {
			step.(*velocity.Setup).Init(
				&backupResolver,
				&j.Build.Task.Commit.Project.Config,
				j.Build.Task.Commit.Hash,
			)
		}
		step.SetProjectRoot(vT.ProjectRoot)
		err := step.Execute(emitter, vT)
		if err != nil {
			break
		}
	}
	wd, _ := os.Getwd()
	if strings.Contains(wd, velocity.WorkspaceDir) {
		os.RemoveAll(wd)
	}
	velocity.GetLogger().Info("completed build", zap.String("buildID", j.Build.ID))
	os.Chdir("/opt/velocityci")
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
