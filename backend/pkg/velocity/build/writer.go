package build

type StreamWriter interface {
	Write(p []byte) (n int, err error)
	SetStatus(s string)
	Close()
}

type StepWriter interface {
	Write(p []byte) (n int, err error)
	SetStatus(s string)
	Close()
}

type TaskWriter interface {
	Write(p []byte) (n int, err error)
	SetStatus(s string)
	Close()
}

// Emitter for forwarding bytes of output onwards
type Emitter interface {
	GetStreamWriter(*Stream) StreamWriter
	GetStepWriter(Step) StepWriter
	GetTaskWriter(*Task) TaskWriter
}

type BlankEmitter struct {
}

func NewBlankEmitter() *BlankEmitter {
	return &BlankEmitter{}
}

func (w *BlankEmitter) GetStreamWriter(*Stream) StreamWriter {
	return &BlankWriter{}
}

func (w *BlankEmitter) GetStepWriter(Step) StepWriter {
	return &BlankWriter{}
}

func (w *BlankEmitter) GetTaskWriter(*Task) TaskWriter {
	return &BlankWriter{}
}

type BlankWriter struct {
}

func (w BlankWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w BlankWriter) SetStatus(s string) {}

func (w BlankWriter) Close() {}
