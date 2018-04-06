package builder

import (
	"os"

	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

func runBuild(build *builder.BuildCtrl, ws *websocket.Conn) {
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
	os.RemoveAll(wd)
	os.Chdir("/opt/velocityci")
}
