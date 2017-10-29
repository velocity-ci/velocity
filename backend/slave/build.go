package main

import (
	"log"
	"os"

	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/api/project"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func runBuild(build *BuildMessage, ws *websocket.Conn) {
	log.Printf("Cloning %s", build.Project.Repository.Address)
	emitter := NewSlaveWriter(
		ws,
		build.Project.ID,
		build.CommitHash,
		build.BuildID,
	)
	emitter.SetTotalSteps(uint64(len(build.Task.Steps) + 1))
	// Cloning is step 0
	repo, dir, err := project.Clone(*build.Project, false, true, emitter)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Done.")
	defer os.RemoveAll(dir)

	w, err := repo.Worktree()
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Printf("Checking out %s", build.CommitHash)
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(build.CommitHash),
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Done.")

	os.Chdir(dir)

	for stepNumber, step := range build.Task.Steps {
		emitter.SetStep(uint64(stepNumber + 1))
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
