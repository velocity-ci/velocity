package builder

import (
	"fmt"

	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
)

// EventPrefix - Global prefix for events.
const EventPrefix = "vlcty_"

// Architect -> Builder events
var (
	EventJobDoPrefix = fmt.Sprintf("%sjob-do-", EventPrefix)
	EventJobStop     = fmt.Sprintf("%sjob-stop", EventPrefix)
)

// Global Builder -> Architect events
var (
	EventJobStatus = fmt.Sprintf("%sjob-status", EventPrefix)
	/*
		{
			"name": "synchronise",
			"status": "running/success/error",
			"errors": [
				{
					"message": ""
				}
			]
		}
	*/

	EventBuilderReady = fmt.Sprintf("%sbuilder-ready", EventPrefix)
	/*
		{}
	*/
)

func SendBuilderReady(ws *phoenix.Client) {
	ws.Socket.Send(&phoenix.PhoenixMessage{
		Event: EventBuilderReady,
		Topic: PoolTopic,
	}, false)
}

type Job interface {
	GetID() string
	GetName() string
	Parse([]byte) error
	Do(*phoenix.Client) error
	Stop(*phoenix.Client) error
}

type baseJob struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (j *baseJob) GetID() string {
	return j.ID
}

func (j *baseJob) GetName() string {
	return j.Name
}

var jobs = []Job{
	NewTask(),
}
