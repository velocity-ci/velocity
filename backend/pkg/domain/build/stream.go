package build

import (
	"encoding/json"
	"time"
)

type Stream struct {
	ID   string `json:"id"`
	Step *Step  `json:"step"`
	Name string `json:"name"`
}

func (s Stream) String() string {
	j, _ := json.Marshal(s)
	return string(j)
}

type StreamLine struct {
	StreamID   string    `json:"streamID"`
	LineNumber int       `json:"lineNumber"`
	Timestamp  time.Time `json:"timestamp"`
	Output     string    `json:"output"`
}
