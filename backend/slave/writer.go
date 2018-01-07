package main

import (
	"log"
	"strings"
	"sync"

	"github.com/velocity-ci/velocity/backend/api/domain/build"

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
	StreamID   string

	LineNumber uint64
	status     string
}

type Emitter struct {
	ws          *safeWebsocket
	BuildStepID string
	Streams     []build.BuildStepStream

	StepNumber uint64
}

func (e *Emitter) NewStreamWriter(streamName string) velocity.StreamWriter {
	streamID := ""
	for _, s := range e.Streams {
		if s.Name == streamName {
			streamID = s.ID
			break
		}
	}
	if streamID == "" {
		log.Fatalf("could not find streamID for %s", streamName)
	}
	return &StreamWriter{
		ws:         e.ws,
		StreamID:   streamID,
		StepNumber: e.StepNumber,
		LineNumber: uint64(0),
	}
}

func (e *Emitter) SetBuildStep(buildStep build.BuildStep) {
	e.BuildStepID = buildStep.ID
	e.Streams = buildStep.Streams
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
		StreamID:   w.StreamID,
		LineNumber: w.LineNumber,
		Status:     w.status,
		Output:     string(p),
	}
	m := slave.SlaveMessage{
		Type: "log",
		Data: lM,
	}
	log.Printf("emitted %s:%d\n%s", w.StreamID, w.LineNumber, p)
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
