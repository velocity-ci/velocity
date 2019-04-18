package builder

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/build"
)

type Emitter struct {
	ws      *phoenix.Client
	BuildID string
	StepID  string
}

func NewEmitter(ws *phoenix.Client) *Emitter {
	return &Emitter{
		ws: ws,
	}
}

func (e *Emitter) GetStreamWriter(stream *build.Stream) build.StreamWriter {
	sw := &BufferedWriter{
		builderBaseWriter: &builderBaseWriter{
			ws:   e.ws,
			open: true,
			// topic: fmt.Sprintf("%sstream:%s", eventBuildPrefix, stream.ID),
			topic:       PoolTopic,
			eventPrefix: fmt.Sprintf("%sstream:%s:", eventBuildPrefix, stream.ID),
		},
		LineNumber: 0,
	}

	go sw.worker()

	return sw
}

func (e *Emitter) GetStepWriter(step build.Step) build.StepWriter {
	sw := &StateWriter{
		builderBaseWriter: &builderBaseWriter{
			ws:    e.ws,
			open:  true,
			topic: fmt.Sprintf("%sstep:%s", eventBuildPrefix, step.GetID()),
		},
	}

	return sw
}

func (e *Emitter) GetTaskWriter(task *build.Task) build.TaskWriter {
	sw := &StateWriter{
		builderBaseWriter: &builderBaseWriter{
			ws:    e.ws,
			open:  true,
			topic: fmt.Sprintf("%stask:%s", eventBuildPrefix, task.ID),
		},
	}

	return sw
}

type builderBaseWriter struct {
	ws          *phoenix.Client
	topic       string
	eventPrefix string
	event       string
	status      string

	open bool
}

func (w *builderBaseWriter) SetStatus(s string) {
	w.status = s
}

func (w *builderBaseWriter) Close() {
	w.open = false
}

type StateWriter struct {
	*builderBaseWriter
}

func (w *StateWriter) Write(p []byte) (n int, err error) {
	w.ws.Socket.Send(&phoenix.PhoenixMessage{
		Event: fmt.Sprintf("%s%s", w.eventPrefix, w.event),
		Topic: w.topic,
		// Payload: map[string]string{},
	}, false)

	return len(p), nil
}

type BufferedWriter struct {
	*builderBaseWriter
	buffer     []*BuildLogLine
	bufferLock sync.Mutex
	LineNumber int
}

func (w *BufferedWriter) worker() {
	for w.open || len(w.buffer) > 0 {
		if len(w.buffer) > 0 {
			w.bufferLock.Lock()
			w.ws.Socket.Send(&phoenix.PhoenixMessage{
				Event: fmt.Sprintf("%s%s", w.eventPrefix, w.event),
				Topic: w.topic,
				Payload: &BuildLogPayload{
					Lines: w.buffer,
				},
			}, false)
			w.buffer = []*BuildLogLine{}
			w.bufferLock.Unlock()
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (w *BufferedWriter) Write(p []byte) (n int, err error) {
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

	w.bufferLock.Lock()
	w.buffer = append(w.buffer, l)
	w.bufferLock.Unlock()

	return len(p), nil
}
