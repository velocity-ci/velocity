package build

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/git"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
)

// Stage represents a set of Tasks that can run in parallel, therefore Tasks is a map[TaskID]Task.
type Stage struct {
	// k: Task.ID
	Tasks map[string]*Task `json:"tasks"`
}

// ConstructionPlan represents a collection of Stages to be executed in order.
type ConstructionPlan struct {
	ID     string   `json:"id"`
	Stages []*Stage `json:"stages"`
}

func NewConstructionPlan(
	targetBlueprintName string,
	blueprints []*config.Blueprint,
	paramResolver BackupResolver,
	repository *git.Repository,
	branch string,
	commitSha string,
	projectRoot string,
) (*ConstructionPlan, error) {
	targetBlueprint, err := getRequestedBlueprintByName(targetBlueprintName, blueprints)
	if err != nil {
		return nil, err
	}
	task := NewTask(
		targetBlueprint,
		paramResolver,
		repository,
		branch,
		commitSha,
		projectRoot,
	)
	return &ConstructionPlan{
		ID: uuid.NewV4().String(),
		Stages: []*Stage{
			&Stage{
				Tasks: map[string]*Task{
					task.ID: task,
				},
			},
		},
	}, nil
}

func (p *ConstructionPlan) Execute(emitter Emitter) error {
	for _, stage := range p.Stages {
		for _, task := range stage.Tasks {
			err := task.Execute(emitter)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getRequestedBlueprintByName(taskName string, tasks []*config.Blueprint) (*config.Blueprint, error) {
	for _, t := range tasks {
		if t.Name == taskName {
			return t, nil
		}
	}

	return nil, fmt.Errorf("could not find %s", taskName)
}
