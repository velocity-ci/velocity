package main

import (
	"github.com/gorilla/websocket"
)

type SlaveWriter struct {
	stepNumber uint64
	totalSteps uint64
	status     string
	ws         *websocket.Conn
	projectID  string
	commitHash string
	buildID    uint64
}

func NewSlaveWriter(ws *websocket.Conn, projectID string, commitHash string, buildID uint64) *SlaveWriter {
	return &SlaveWriter{
		ws:         ws,
		projectID:  projectID,
		commitHash: commitHash,
		buildID:    buildID,
	}
}

func (w SlaveWriter) Write(p []byte) (n int, err error) {
	lM := LogMessage{
		ProjectID:  w.projectID,
		CommitHash: w.commitHash,
		BuildID:    w.buildID,
		Step:       w.stepNumber,
		Status:     w.status,
		Output:     string(p),
	}
	m := SlaveMessage{
		Type: "log",
		Data: lM,
	}
	err = w.ws.WriteJSON(m)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *SlaveWriter) SetStep(num uint64) {
	w.stepNumber = num
}

func (w *SlaveWriter) SetStatus(s string) {
	w.status = s
}

func (w *SlaveWriter) SetTotalSteps(t uint64) {
	w.totalSteps = t
}
