package main

import (
	"encoding/json"
	"fmt"

	"github.com/velocity-ci/velocity/backend/api/project"
	"github.com/velocity-ci/velocity/backend/task"
)

type CommandMessage struct {
	Command string  `json:"command"`
	Data    Message `json:"data"`
}

type Message interface{}

type BuildMessage struct {
	Project    *project.Project `json:"project"`
	CommitHash string           `json:"commit"`
	BuildID    uint64           `json:"buildId"`
	Task       *task.Task       `json:"task"`
}

func (c *CommandMessage) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	// We'll store the error (if any) so we can return it if necessary
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	// Deserialize Command
	err = json.Unmarshal(*objMap["command"], &c.Command)
	if err != nil {
		return err
	}

	// Deserialize Data by command
	var rawData json.RawMessage
	err = json.Unmarshal(*objMap["data"], &rawData)
	if err != nil {
		return err
	}

	if c.Command == "build" {
		d := BuildMessage{}
		err := json.Unmarshal(rawData, &d)
		if err != nil {
			return err
		}
		c.Data = &d
	} else {
		return fmt.Errorf("unsupported type in json.Unmarshal: %s", c.Command)
	}

	return nil
}
