package main

import (
	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/api/slave"
	"github.com/velocity-ci/velocity/backend/velocity"
)

func runBuild(build *slave.BuildCommand, ws *websocket.Conn) {
	emitter := NewSlaveWriter(ws)

	for i, step := range build.Task.Steps {
		emitter.SetBuildStepID(build.Build.Steps[i].ID)

		step.SetParams(build.Build.Parameters)
		if step.GetType() == "clone" {
			step.(*velocity.Clone).SetGitRepositoryAndCommitHash(
				build.Project.Repository,
				build.Commit.Hash,
			)
		}

		err := step.Execute(emitter, build.Build.Parameters)
		if err != nil {
			break
		}
	}
}
