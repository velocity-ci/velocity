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
	ws *phoenix.Client
}

func NewEmitter(ws *phoenix.Client) *Emitter {
	return &Emitter{
		ws: ws,
	}
}

func (e *Emitter) GetStreamWriter(stream *build.Stream) build.StreamWriter {
	sw := &BufferedWriter{
		builderBaseWriter: &builderBaseWriter{
			ws:    e.ws,
			open:  true,
			topic: PoolTopic,
			event: fmt.Sprintf("%sstream:new-loglines", eventBuildPrefix),
		},
		bufferedPayloadRenderer: func(w *BufferedWriter) interface{} {
			return &StreamNewLogLinePayload{
				ID:    stream.ID,
				Lines: w.buffer,
			}
		},
		bufferedRenderer: func(w *BufferedWriter, p []byte) interface{} {
			o := strings.TrimSpace(string(p))
			if !strings.HasSuffix(string(p), "\r") {
				w.lineNumber++
				o += "\n"
			}

			l := &BuildLogLine{
				Timestamp:  time.Now().UTC(),
				LineNumber: w.lineNumber,
				Output:     o,
			}

			return l
		},
	}

	go sw.worker()

	return sw
}

func (e *Emitter) GetStepWriter(step build.Step) build.StepWriter {
	sw := &StateWriter{
		builderBaseWriter: &builderBaseWriter{
			ws:    e.ws,
			open:  true,
			topic: PoolTopic,
			event: fmt.Sprintf("%sstep:update", eventBuildPrefix),
			payloadRenderer: func(w *builderBaseWriter, p []byte) interface{} {
				return &StepUpdatePayload{
					ID:    step.GetID(),
					State: w.status,
				}
			},
		},
	}

	return sw
}

func (e *Emitter) GetTaskWriter(task *build.Task) build.TaskWriter {
	sw := &StateWriter{
		builderBaseWriter: &builderBaseWriter{
			ws:    e.ws,
			open:  true,
			topic: PoolTopic,
			event: fmt.Sprintf("%stask:update", eventBuildPrefix),
			payloadRenderer: func(w *builderBaseWriter, p []byte) interface{} {
				return &TaskUpdatePayload{
					ID:    task.ID,
					State: w.status,
				}
			},
		},
	}

	return sw
}

type builderBaseWriter struct {
	ws              *phoenix.Client
	topic           string
	event           string
	status          string
	payloadRenderer func(*builderBaseWriter, []byte) interface{}

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
	id string
}

func (w *StateWriter) Write(p []byte) (n int, err error) {
	w.ws.Socket.Send(&phoenix.PhoenixMessage{
		Event:   w.event,
		Topic:   w.topic,
		Payload: w.payloadRenderer(w.builderBaseWriter, p),
	}, false)

	return len(p), nil
}

type BufferedWriter struct {
	*builderBaseWriter
	buffer     []interface{}
	bufferLock sync.Mutex

	bufferedRenderer        func(*BufferedWriter, []byte) interface{}
	bufferedPayloadRenderer func(*BufferedWriter) interface{}

	lineNumber int
	id         string
}

func (w *BufferedWriter) worker() {
	for w.open || len(w.buffer) > 0 {
		if len(w.buffer) > 0 {
			w.bufferLock.Lock()
			w.ws.Socket.Send(&phoenix.PhoenixMessage{
				Event:   w.event,
				Topic:   w.topic,
				Payload: w.bufferedPayloadRenderer(w),
			}, false)
			w.buffer = []interface{}{}
			w.bufferLock.Unlock()
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (w *BufferedWriter) Write(p []byte) (n int, err error) {

	l := w.bufferedRenderer(w, p)

	w.bufferLock.Lock()
	w.buffer = append(w.buffer, l)
	w.bufferLock.Unlock()

	return len(p), nil
}
