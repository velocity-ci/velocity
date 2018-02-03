package build

import (
	"encoding/json"
	"log"
	"time"

	"github.com/velocity-ci/velocity/backend/velocity"
)

type Step struct {
	UUID string `json:"id"`
	// Build  *Build `json:"build"`
	Number int `json:"number"`

	VStep *velocity.Step `json:"step"`

	Streams []*Stream `json:"streams"`

	Status      string    `json:"status"` // waiting, running, success, failed
	UpdatedAt   time.Time `json:"updatedAt"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt"`
}

func (s *Step) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	json.Unmarshal(*objMap["id"], &s.UUID)
	// json.Unmarshal(*objMap["build"], &s.Build)
	json.Unmarshal(*objMap["number"], &s.Number)
	json.Unmarshal(*objMap["status"], &s.Status)
	json.Unmarshal(*objMap["updatedAt"], &s.UpdatedAt)
	json.Unmarshal(*objMap["startedAt"], &s.StartedAt)
	json.Unmarshal(*objMap["completedAt"], &s.CompletedAt)

	var rawStep *json.RawMessage
	err = json.Unmarshal(*objMap["step"], &rawStep)
	if err == nil {
		var m map[string]interface{}
		err = json.Unmarshal(*rawStep, &m)
		if err != nil {
			log.Println("could not unmarshal step")
			return err
		}

		step, err := velocity.DetermineStepFromInterface(m)
		if err != nil {
			log.Println(err)
		} else {
			err := json.Unmarshal(*rawStep, step)
			if err != nil {
				log.Println(err)
			} else {
				s.VStep = &step
			}
		}
	}

	s.Streams = []*Stream{}
	err = json.Unmarshal(*objMap["streams"], &s.Streams)
	if err != nil {
		return err
	}

	return nil
}
