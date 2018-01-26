package main

import (
	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/api/slave"
	"github.com/velocity-ci/velocity/backend/velocity"
)

func runBuild(build *slave.BuildCommand, ws *websocket.Conn) {
	emitter := NewEmitter(ws)

	backupResolver := NewParameterResolver(build.Build.Parameters)

	t := &build.Task.Task

	for _, step := range build.Build.Steps {
		emitter.SetBuildStep(step)

		if step.VStep.GetType() == "setup" {
			step.VStep.(*velocity.Setup).Init(
				&backupResolver,
				&build.Project.Repository,
				build.Commit.Hash)
		}

		err := step.VStep.Execute(emitter, t)
		if err != nil {
			break
		}
	}
}
