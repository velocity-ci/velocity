package main

import (
	"log"
	"strings"
	"sync"

	"github.com/velocity-ci/velocity/backend/velocity"

	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/api/slave"
)

type safeWebsocket struct {
	ws   *websocket.Conn
	lock sync.RWMutex
}

func (sws *safeWebsocket) WriteJSON(m interface{}) error {
	sws.lock.Lock()
	defer sws.lock.Unlock()

	return sws.ws.WriteJSON(m)
}

type StreamWriter struct {
	ws         *safeWebsocket
	StepNumber uint64
	StreamName string

	BuildStepID string
	LineNumber  uint64
	status      string
}

type Emitter struct {
	ws          *safeWebsocket
	BuildStepID string
	StepNumber  uint64
}

func (e *Emitter) NewStreamWriter(streamName string) velocity.StreamWriter {
	return &StreamWriter{
		ws:         e.ws,
		StreamName: streamName,
		StepNumber: e.StepNumber,
		LineNumber: uint64(0),
	}
}

func (e *Emitter) SetBuildStepID(buildStepID string) {
	e.BuildStepID = buildStepID
}

func (e *Emitter) SetStepNumber(n uint64) {
	e.StepNumber = n
}

func NewEmitter(ws *websocket.Conn) *Emitter {
	return &Emitter{
		ws: &safeWebsocket{ws: ws},
	}
}

func (w *StreamWriter) Write(p []byte) (n int, err error) {
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
	log.Printf("emitted %s:%s:%d\n%s", w.BuildStepID, w.StreamName, w.LineNumber, p)
	err = w.ws.WriteJSON(m)
	if err != nil {
		return 0, err
	}
	if !strings.ContainsRune(string(p), '\r') {
		w.LineNumber++
	}
	return len(p), nil
}

func (w *StreamWriter) SetStatus(s string) {
	w.status = s
}
