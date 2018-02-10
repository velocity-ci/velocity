package builder

import (
	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

func runBuild(build *builder.BuildCtrl, ws *websocket.Conn) {
	emitter := NewEmitter(ws, build.Build)

	backupResolver := NewParameterResolver(build.Build.Parameters)

	t := build.Build.Task

	for _, step := range build.Steps {
		emitter.SetStepAndStreams(step, build.Streams)

		s := *step.VStep
		if s.GetType() == "setup" {
			s.(*velocity.Setup).Init(
				&backupResolver,
				&build.Build.Task.Commit.Project.Config,
				build.Build.Task.Commit.Hash,
			)
		}

		err := s.Execute(emitter, t.VTask)
		if err != nil {
			break
		}
	}
}
