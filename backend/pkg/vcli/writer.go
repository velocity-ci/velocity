package vcli

import (
	"fmt"
	"strings"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/build"
)

type Emitter struct {
	StepNumber uint64
}

func NewEmitter() *Emitter {
	return &Emitter{}
}

func (e *Emitter) GetStreamWriter(stream *build.Stream) build.StreamWriter {
	return &StdOutWriter{
		prefix: fmt.Sprintf("%s:  ", stream.Name),
	}
}

func (e *Emitter) GetStepWriter(step build.Step) build.StepWriter {
	return &StdOutWriter{
		prefix: "",
	}
}

func (e *Emitter) GetTaskWriter(task *build.Task) build.TaskWriter {
	return &StdOutWriter{
		prefix: "",
	}
}

type StdOutWriter struct {
	prefix     string
	status     string
	ansiColour string
}

func (w *StdOutWriter) Write(p []byte) (n int, err error) {
	fmt.Printf("%s%s", w.prefix, string(p))
	if !strings.HasSuffix(string(p), "\r") {
		fmt.Println()
	}
	return len(p), nil
}

func (w *StdOutWriter) SetStatus(s string) {
	w.status = s
}

func (w *StdOutWriter) Close() {
	return
}
