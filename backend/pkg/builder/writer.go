package builder

import (
	"strings"
	"sync"

	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
	"go.uber.org/zap"

	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
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
	StepNumber int

	BuildID  string
	StepID   string
	StreamID string

	LineNumber int
	status     string
}

type Emitter struct {
	ws      *safeWebsocket
	BuildID string
	StepID  string
	Streams []*build.Stream

	StepNumber int
}

func (e *Emitter) GetStreamWriter(streamName string) velocity.StreamWriter {
	streamID := ""
	for _, s := range e.Streams {
		if s.Name == streamName {
			streamID = s.ID
			break
		}
	}
	if streamID == "" {
		velocity.GetLogger().Error("could not find streamID", zap.String("stream name", streamName))
	}
	return &StreamWriter{
		ws:         e.ws,
		BuildID:    e.BuildID,
		StepID:     e.StepID,
		StreamID:   streamID,
		StepNumber: e.StepNumber,
		LineNumber: 0,
	}
}

func (e *Emitter) SetStepAndStreams(step *build.Step, streams []*build.Stream) {
	e.StepID = step.ID
	e.Streams = []*build.Stream{}
	for _, s := range streams {
		if s.Step.ID == step.ID {
			e.Streams = append(e.Streams, s)
		}
	}
}

func (e *Emitter) SetStepNumber(n int) {
	e.StepNumber = n
}

func NewEmitter(ws *websocket.Conn, b *build.Build) *Emitter {
	return &Emitter{
		ws:      &safeWebsocket{ws: ws},
		BuildID: b.ID,
	}
}

func (w *StreamWriter) Write(p []byte) (n int, err error) {
	o := strings.TrimSpace(string(p))
	if !strings.HasSuffix(string(p), "\r") {
		w.LineNumber++
		o += "\n"
	}

	lM := builder.BuilderStreamLineMessage{
		BuildID:    w.BuildID,
		StepID:     w.StepID,
		StreamID:   w.StreamID,
		LineNumber: w.LineNumber,
		Status:     w.status,
		Output:     o,
	}
	m := builder.BuilderRespMessage{
		Type: "log",
		Data: lM,
	}
	err = w.ws.WriteJSON(m)

	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (w *StreamWriter) SetStatus(s string) {
	w.status = s
}
