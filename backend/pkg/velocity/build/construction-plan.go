package build

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/git"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
)

// Stage represents a set of Tasks that can run in parallel, therefore Tasks is a map[TaskID]Task.
type Stage struct {
	ID     string `json:"id"`
	Index  uint16 `json:"index"`
	Status string `json:"status"`
	// k: Task.ID
	Tasks map[string]*Task `json:"tasks"`
}

// ConstructionPlan represents a collection of Stages to be executed in order.
type ConstructionPlan struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Stages []*Stage `json:"stages"`
	// Plugins []*Plugin `json:"plugins"`
}

func NewConstructionPlanFromBlueprint(
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
		ID:   uuid.NewV4().String(),
		Name: fmt.Sprintf("Blueprint: %s", targetBlueprintName),
		Stages: []*Stage{
			{
				ID:     uuid.NewV4().String(),
				Index:  1,
				Status: StateWaiting,
				Tasks: map[string]*Task{
					task.ID: task,
				},
			},
		},
	}, nil
}

func NewConstructionPlanFromPipeline(
	targetPipelineName string,
	pipelines []*config.Pipeline,
	blueprints []*config.Blueprint,
	paramResolver BackupResolver,
	repository *git.Repository,
	branch string,
	commitSha string,
	projectRoot string,
) (*ConstructionPlan, error) {
	targetPipeline, err := getRequestedPipelineByName(targetPipelineName, pipelines)
	if err != nil {
		return nil, err
	}

	cP := &ConstructionPlan{
		ID:     uuid.NewV4().String(),
		Name:   fmt.Sprintf("Pipeline: %s", targetPipelineName),
		Stages: []*Stage{},
	}

	for i, stage := range targetPipeline.Stages {
		newStage := &Stage{
			ID:     uuid.NewV4().String(),
			Index:  uint16(i + 1),
			Status: StateWaiting,
			Tasks:  map[string]*Task{},
		}
		for _, blueprintName := range stage.Blueprints {
			blueprint, err := getRequestedBlueprintByName(blueprintName, blueprints)
			if err != nil {
				return nil, err
			}
			newTask := NewTask(
				blueprint,
				paramResolver,
				repository,
				branch,
				commitSha,
				projectRoot,
			)
			newStage.Tasks[newTask.ID] = newTask
		}
		cP.Stages = append(cP.Stages, newStage)
	}

	return cP, nil
}

func (p *ConstructionPlan) Execute(emitter Emitter) error {
	eventBuildStart(p)
	defer eventBuildComplete(p)
	for _, stage := range p.Stages {
		for _, task := range stage.Tasks {
			eventTaskStart(p, task)
			err := task.Execute(emitter)
			eventTaskComplete(p, task)
			if err != nil {
				eventTaskFail(p, task, err)
				if !task.IgnoreErrors {
					eventBuildFail(p, task, err)
					return err
				}
			}
			eventTaskSuccess(p, task)
		}
	}

	eventBuildSuccess(p)

	return nil
}

func (p *ConstructionPlan) GracefulStop() error {
	for _, stage := range p.Stages {
		for _, task := range stage.Tasks {
			err := task.GracefulStop()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getRequestedBlueprintByName(blueprintName string, blueprints []*config.Blueprint) (*config.Blueprint, error) {
	for _, b := range blueprints {
		if b.Name == blueprintName {
			return b, nil
		}
	}

	return nil, fmt.Errorf("could not find %s", blueprintName)
}

func getRequestedPipelineByName(pipelineName string, pipelines []*config.Pipeline) (*config.Pipeline, error) {
	for _, p := range pipelines {
		if p.Name == pipelineName {
			return p, nil
		}
	}

	return nil, fmt.Errorf("could not find %s", pipelineName)
}
