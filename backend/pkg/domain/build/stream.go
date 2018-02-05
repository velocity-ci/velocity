package build

import (
	"time"
)

type Stream struct {
	ID string `json:"id"`
	// Step *Step  `json:"step"`
	Name string `json:"name"`
}

type StreamLine struct {
	StreamID   string    `json:"streamID"`
	LineNumber int       `json:"lineNumber"`
	Timestamp  time.Time `json:"timestamp"`
	Output     string    `json:"output"`
}
