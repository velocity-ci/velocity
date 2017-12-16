package main

import "fmt"

type CLIWriter struct {
}

func NewCLIWriter() *CLIWriter {
	return &CLIWriter{}
}

func (w CLIWriter) Write(p []byte) (n int, err error) {
	fmt.Printf("    %s\n", p)
	return len(p), nil
}

func (w *CLIWriter) SetStatus(s string) {
}

func (w *CLIWriter) SetStreamName(n string) {
}
