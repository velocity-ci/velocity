package main

import (
	"encoding/json"
	"fmt"

	"github.com/velocity-ci/velocity/backend/api/slave"
)

type CommandMessage struct {
	Command string  `json:"command"`
	Data    Command `json:"data"`
}

type Command interface{}

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
		d := slave.BuildCommand{}
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
