package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/api/project"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func runBuild(build *BuildMessage, ws *websocket.Conn) {
	repo, dir, err := project.Clone(*build.Project, false)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer os.RemoveAll(dir)

	w, err := repo.Worktree()
	if err != nil {
		log.Fatal(err)
		return
	}
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(build.CommitHash),
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Checked out %s", build.CommitHash)
	s, _ := w.Status()
	log.Printf(s.String())

	os.Chdir(dir)
	files, err := ioutil.ReadDir("./")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Println(f.Name())
	}

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
