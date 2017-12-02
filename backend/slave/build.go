package main

import (
	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/api/slave"
)

func runBuild(build *slave.BuildCommand, ws *websocket.Conn) {
	emitter := NewSlaveWriter(ws)

	for _, buildStep := range build.BuildSteps {
		buildStep.SetParams(build.Build.Parameters)
		if buildStep.GetType() == "clone" {
			// buildStep.Step.(*velocity.Clone).SetBuild(velocity.NewBuild(build.Project, build.CommitHash, build.BuildID))
		}

		emitter.SetStatus("running")
		err := buildStep.Execute(emitter, build.Build.Parameters)
		if err != nil {
			break
		}
	}
}
