package builder

import (
	"encoding/json"
	"fmt"

	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"
)

type BuilderCtrlMessage struct {
	Command string      `json:"command"`
	Payload interface{} `json:"payload"`
}

type BuildCtrl struct {
	Build   *build.Build    `json:"build"`
	Steps   []*build.Step   `json:"steps"`
	Streams []*build.Stream `json:"streams"`
}

func newBuildCommand(b *build.Build, steps []*build.Step, streams []*build.Stream) *BuilderCtrlMessage {
	return &BuilderCtrlMessage{
		Command: "build",
		Payload: &BuildCtrl{
			Build:   b,
			Steps:   steps,
			Streams: streams,
		},
	}
}

type KnownHostCtrl struct {
	KnownHosts []*knownhost.KnownHost `json:"knownHosts"`
}

func newKnownHostsCommand(ks []*knownhost.KnownHost) *BuilderCtrlMessage {
	return &BuilderCtrlMessage{
		Command: "knownhosts",
		Payload: &KnownHostCtrl{
			KnownHosts: ks,
		},
	}
}

func (c *BuilderCtrlMessage) UnmarshalJSON(b []byte) error {
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
	err = json.Unmarshal(*objMap["payload"], &rawData)
	if err != nil {
		return err
	}

	if c.Command == "build" {
		d := BuildCtrl{}
		err := json.Unmarshal(rawData, &d)
		if err != nil {
			return err
		}
		c.Payload = &d
	} else if c.Command == "knownhosts" {
		d := KnownHostCtrl{}
		err := json.Unmarshal(rawData, &d)
		if err != nil {
			return err
		}
		c.Payload = &d
	} else {
		return fmt.Errorf("unsupported type in json.Unmarshal: %s", c.Command)
	}

	return nil
}

type BuilderStreamLineMessage struct {
	BuildID    string `json:"buildId"`
	StepID     string `json:"stepId"`
	StreamID   string `json:"streamId"`
	LineNumber int    `json:"lineNumber"`
	Status     string `json:"status"`
	Output     string `json:"output"`
}

type BuilderRespMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (c *BuilderRespMessage) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	// We'll store the error (if any) so we can return it if necessary
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	// Deserialize Command
	err = json.Unmarshal(*objMap["type"], &c.Type)
	if err != nil {
		return err
	}

	// Deserialize Data by command
	var rawData json.RawMessage
	err = json.Unmarshal(*objMap["data"], &rawData)
	if err != nil {
		return err
	}

	if c.Type == "log" {
		d := BuilderStreamLineMessage{}
		err := json.Unmarshal(rawData, &d)
		if err != nil {
			return err
		}
		c.Data = &d
	} else {
		return fmt.Errorf("unsupported type in json.Unmarshal: %s", c.Type)
	}

	return nil
}
