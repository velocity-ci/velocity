package io

import (
	"fmt"
)

type StreamWriter interface {
	Write(p []byte) (n int, err error)
	SetStatus(s string)
	Close()
}

// Emitter for forwarding bytes of output onwards
type Emitter interface {
	GetStreamWriter(streamName string) StreamWriter
}

type BlankEmitter struct {
}

func NewBlankEmitter() *BlankEmitter {
	return &BlankEmitter{}
}

func (w *BlankEmitter) GetStreamWriter(streamName string) StreamWriter {
	return &BlankWriter{}
}

type BlankWriter struct {
}

func (w BlankWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w BlankWriter) SetStatus(s string) {}

func (w BlankWriter) Close() {}

const (
	ANSISuccess = "\x1b[1m\x1b[49m\x1b[32m"
	ANSIWarn    = "\x1b[1m\x1b[49m\x1b[33m"
	ANSIError   = "\x1b[1m\x1b[49m\x1b[31m"
	ANSIInfo    = "\x1b[1m\x1b[49m\x1b[34m"
)

func ColorFmt(ansiColor, format string) string {
	return fmt.Sprintf("%s%s\x1b[0m", ansiColor, format)
}
