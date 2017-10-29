package main

import "fmt"

type CLIWriter struct {
	stepNumber uint64
	totalSteps uint64
	status     string
}

func NewCLIWriter() *CLIWriter {
	return &CLIWriter{}
}

func (w CLIWriter) Write(p []byte) (n int, err error) {
	fmt.Printf("    %s\n", p)
	return len(p), nil
}

func (w *CLIWriter) SetStep(num uint64) {
	w.stepNumber = num
}

func (w *CLIWriter) SetStatus(s string) {
	w.status = s
}

func (w *CLIWriter) SetTotalSteps(t uint64) {
	w.totalSteps = t
}
