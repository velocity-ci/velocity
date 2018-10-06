package builder

import (
	"os"
	"strings"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

func (b *Builder) runBuild(build *BuildPayload) {
	velocity.GetLogger().Info("running build", zap.String("buildID", build.Build.ID))

	emitter := NewEmitter(b.ws, build.Build)

	backupResolver := NewParameterResolver(build.Build.Parameters)

	vT := build.Build.Task.VTask

	for i, step := range vT.Steps {
		bStep := build.Steps[i]
		velocity.GetLogger().Debug("running step", zap.String("stepID", bStep.ID))
		emitter.SetStepAndStreams(bStep, build.Streams)

		if step.GetType() == "setup" {
			step.(*velocity.Setup).Init(
				&backupResolver,
				&build.Build.Task.Commit.Project.Config,
				build.Build.Task.Commit.Hash,
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
	velocity.GetLogger().Info("completed build", zap.String("buildID", build.Build.ID))
	os.Chdir("/opt/velocityci")
}

type BuildPayload struct {
	Build   *build.Build    `json:"build"`
	Steps   []*build.Step   `json:"steps"`
	Streams []*build.Stream `json:"streams"`
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
