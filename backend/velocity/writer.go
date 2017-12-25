package velocity

type StreamWriter interface {
	Write(p []byte) (n int, err error)
	SetStatus(s string)
}

// Emitter for forwarding bytes of output onwards
type Emitter interface {
	NewStreamWriter(streamName string) StreamWriter
}

type BlankEmitter struct {
}

func NewBlankEmitter() *BlankEmitter {
	return &BlankEmitter{}
}

func (w *BlankEmitter) NewStreamWriter(streamName string) StreamWriter {
	return &BlankWriter{}
}

type BlankWriter struct {
}

func (w BlankWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w BlankWriter) SetStatus(s string) {}
