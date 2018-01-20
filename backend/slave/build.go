package main

import (
	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/api/slave"
	"github.com/velocity-ci/velocity/backend/velocity"
)

func runBuild(build *slave.BuildCommand, ws *websocket.Conn) {
	emitter := NewEmitter(ws)

	backupResolver := NewParameterResolver(build.Build.Parameters)
	build.Task.Setup(
		emitter,
		&backupResolver,
		&build.Project.Repository,
		build.Commit.Hash,
	)

	for i, step := range build.Task.Steps {
		emitter.SetBuildStep(build.Build.Steps[i])

		err := step.Execute(emitter, map[string]velocity.Parameter{})
		if err != nil {
			break
		}
	}
}
