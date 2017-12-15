package velocity

// Emitter for forwarding bytes of output onwards
type Emitter interface {
	Write(p []byte) (n int, err error)
	SetStreamName(name string)
	SetStatus(s string)
}

type BlankWriter struct {
}

func NewBlankWriter() *BlankWriter {
	return &BlankWriter{}
}

func (w BlankWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w *BlankWriter) SetStreamName(name string) {}

func (w *BlankWriter) SetStatus(s string) {}
