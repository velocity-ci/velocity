package main

import (
	"log"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/api/slave"
)

type SlaveWriter struct {
	ws          *websocket.Conn
	BuildStepID string
	StreamName  string

	LineNumber uint64
	status     string
}

func NewSlaveWriter(ws *websocket.Conn) *SlaveWriter {
	return &SlaveWriter{
		ws: ws,
	}
}

func (w *SlaveWriter) Write(p []byte) (n int, err error) {
	lM := slave.SlaveBuildLogMessage{
		BuildStepID: w.BuildStepID,
		StreamName:  w.StreamName,
		LineNumber:  w.LineNumber,
		Status:      w.status,
		Output:      string(p),
	}
	m := slave.SlaveMessage{
		Type: "log",
		Data: lM,
	}
	log.Printf("emitted %s:%s:%d:%s\n%s", w.BuildStepID, w.StreamName, w.LineNumber, w.status, p)
	err = w.ws.WriteJSON(m)
	if err != nil {
		return 0, err
	}
	if !strings.ContainsRune(string(p), '\r') {
		w.LineNumber++
	}
	return len(p), nil
}

func (w *SlaveWriter) SetLineNumber(num uint64) {
	w.LineNumber = num
}

func (w *SlaveWriter) SetStatus(s string) {
	w.status = s
}

func (w *SlaveWriter) SetBuildStepID(id string) {
	w.BuildStepID = id
}

func (w *SlaveWriter) SetStreamName(name string) {
	w.StreamName = name
	w.LineNumber = 0
}
