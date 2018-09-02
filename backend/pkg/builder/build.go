package builder

import (
	"os"
	"strings"

	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

func runBuild(build *builder.BuildCtrl, ws *phoenix.PhoenixWSClient) {
	emitter := NewEmitter(ws, build.Build)

	backupResolver := NewParameterResolver(build.Build.Parameters)

	vT := build.Build.Task.VTask

	for i, step := range vT.Steps {
		bStep := build.Steps[i]
		emitter.SetStepAndStreams(bStep, build.Streams)

		// s := *step.VStep
		if step.GetType() == "setup" {
			step.(*velocity.Setup).Init(
				&backupResolver,
				&build.Build.Task.Commit.Project.Config,
				build.Build.Task.Commit.Hash,
			)
		}

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
