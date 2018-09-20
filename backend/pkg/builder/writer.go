package builder

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
	"go.uber.org/zap"

	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

// TODO: Debouncing
type StreamWriter struct {
	ws         *phoenix.Client
	StepNumber int

	BuildID  string
	StepID   string
	StreamID string

	LineNumber int
	status     string

	buffer     []*BuildLogLine
	bufferLock sync.Mutex

	open bool
}

type Emitter struct {
	ws      *phoenix.Client
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

	sw := &StreamWriter{
		ws:         e.ws,
		BuildID:    e.BuildID,
		StepID:     e.StepID,
		StreamID:   streamID,
		StepNumber: e.StepNumber,
		LineNumber: 0,
		open:       true,
	}

	go sw.worker()

	return sw
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

func NewEmitter(ws *phoenix.Client, b *build.Build) *Emitter {
	return &Emitter{
		ws:      ws,
		BuildID: b.ID,
	}
}

func (w *StreamWriter) Write(p []byte) (n int, err error) {
	o := strings.TrimSpace(string(p))
	if !strings.HasSuffix(string(p), "\r") {
		w.LineNumber++
		o += "\n"
	}

	l := &BuildLogLine{
		Timestamp:  time.Now().UTC(),
		LineNumber: w.LineNumber,
		Status:     w.status,
		Output:     o,
	}

	w.buffer = append(w.buffer, l)

	return len(p), nil
}

func (w *StreamWriter) SetStatus(s string) {
	w.status = s
}

func (w *StreamWriter) Close() {
	w.open = false
}

func (w *StreamWriter) worker() {
	for w.open || len(w.buffer) > 0 {
		if len(w.buffer) > 0 {
			w.bufferLock.Lock()
			w.ws.Socket.Send(&phoenix.PhoenixMessage{
				Event: EventNewStreamLines,
				Topic: fmt.Sprintf("stream:%s", w.StreamID),
				Payload: &BuildLogPayload{
					StreamID: w.StreamID,
					Lines:    w.buffer,
				},
			}, false)
			w.buffer = []*BuildLogLine{}
			w.bufferLock.Unlock()
			time.Sleep(500 * time.Millisecond)
		}
	}
}

const (
	EventNewStreamLines = "streamLines:new"
)
