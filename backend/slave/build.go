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
	repo, dir, err := project.Clone(*build.Project, false, true)
	if err != nil {
		log.Fatalf("18: %v", err)
		return
	}
	log.Println("Done.")
	defer os.RemoveAll(dir)

	w, err := repo.Worktree()
	if err != nil {
		log.Fatalf("25: %v", err)
		return
	}
	log.Printf("Checking out %s", build.CommitHash)
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(build.CommitHash),
	})
	if err != nil {
		log.Fatalf("32: %v", err)
		return
	}
	log.Println("Done.")

	os.Chdir(dir)

	emit := func(status string, step uint64, output string) {
		lM := LogMessage{
			ProjectID:  build.Project.ID,
			CommitHash: build.CommitHash,
			BuildID:    build.BuildID,
			Step:       step,
			Status:     status,
			Output:     output,
		}
		m := SlaveMessage{
			Type: "log",
			Data: lM,
		}
		err := ws.WriteJSON(m)
		if err != nil {
			log.Fatal(err)
		}
		// log.Printf("Emitted. Project: %s, Commit: %s, BuildID: %d Step: %d, Status: %s", lM.ProjectID, lM.CommitHash, lM.BuildID, lM.Step, lM.Status)
	}

	build.Task.SetEmitter(emit)
	os.Chdir(dir)
	for stepNumber, step := range build.Task.Steps {
		err := step.Execute(uint64(stepNumber), build.Task.Parameters)
		if err != nil {
			break
		}
		emit("success", uint64(stepNumber), "")
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
