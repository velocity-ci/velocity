package velocity

// Emitter for forwarding bytes of output onwards
type Emitter interface {
	Write(p []byte) (n int, err error)
	SetStep(num uint64)
	SetStatus(s string)
	SetTotalSteps(t uint64)
}

type BlankWriter struct {
}

func NewBlankWriter() *BlankWriter {
	return &BlankWriter{}
}

func (w BlankWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w *BlankWriter) SetStep(num uint64) {}

func (w *BlankWriter) SetStatus(s string) {}

func (w *BlankWriter) SetTotalSteps(t uint64) {}
