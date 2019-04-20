package build

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/git"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
)

type Stage struct {
	// k: Task.ID
	Tasks map[string]*Task `json:"tasks"`
}

type ConstructionPlan struct {
	ID     string   `json:"id"`
	Stages []*Stage `json:"stages"`
}

func NewConstructionPlan(
	blueprintName string,
	blueprints []*config.Blueprint,
	paramResolver BackupResolver,
	repository *git.Repository,
	branch string,
	commitSha string,
	projectRoot string,
) (*ConstructionPlan, error) {
	blueprint, err := getRequestedBlueprintByName(blueprintName, blueprints)
	if err != nil {
		return nil, err
	}
	task := NewTask(
		blueprint,
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
