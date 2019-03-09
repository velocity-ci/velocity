package vcli

import (
	"fmt"
	"strings"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
)

type StreamWriter struct {
	StreamName string
	status     string
	ansiColour string
}

type Emitter struct {
	StepNumber uint64
}

func NewEmitter() *Emitter {
	return &Emitter{}
}

func (e *Emitter) GetStreamWriter(streamName string) out.StreamWriter {
	return &StreamWriter{
		StreamName: streamName,
	}
}

func (w *StreamWriter) Write(p []byte) (n int, err error) {
	fmt.Printf("%s:    %s", w.StreamName, string(p))
	if !strings.HasSuffix(string(p), "\r") {
		fmt.Println()
	}
	return len(p), nil
}

func (w *StreamWriter) SetStatus(s string) {
	w.status = s
}

func (w *StreamWriter) Close() {
	return
}
