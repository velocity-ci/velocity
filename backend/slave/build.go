package main

import (
	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/velocity"
)

func runBuild(build *BuildMessage, ws *websocket.Conn) {
	emitter := NewSlaveWriter(
		ws,
		build.Project.ID,
		build.CommitHash,
		build.BuildID,
	)

	for stepNumber, step := range build.Task.Steps {
		if step.GetType() == "clone" {
			step.(*velocity.Clone).SetBuild(velocity.NewBuild(build.Project, build.CommitHash, build.BuildID))
		}

		emitter.SetStep(uint64(stepNumber))
		emitter.SetStatus("running")
		err := step.Execute(emitter, build.Task.Parameters)
		if err != nil {
			break
		}
	}
}

type SlaveMessage struct {
	Type string  `json:"type"`
	Data Message `json:"data"`
}

type LogMessage struct {
	ProjectID  string `json:"projectID"`
	CommitHash string `json:"commitHash"`
	BuildID    uint64 `json:"buildId"`
	Step       uint64 `json:"step"`
	Status     string `json:"status"`
	Output     string `json:"output"`
}
