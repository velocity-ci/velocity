package build

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/git"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
)

type Task struct {
	ID         string `json:"id"`
	parameters map[string]*Parameter

	Blueprint    config.Blueprint `json:"blueprint"`
	IgnoreErrors bool             `json:"ignoreErrors"`
	Docker       TaskDocker       `json:"docker"`
	Steps        []Step           `json:"steps"`

	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"startedAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
	CompletedAt *time.Time `json:"completedAt"`

	ProjectRoot string `json:"-"`
}

func (t *Task) UnmarshalJSON(b []byte) error {
	// We don't return any errors from this function so we can show more helpful parse errors
	var objMap map[string]*json.RawMessage
	// We'll store the error (if any) so we can return it if necessary
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	// Deserialize Blueprint
	err = json.Unmarshal(*objMap["blueprint"], &t.Blueprint)
	if err != nil {
		return err
	}

	// Deserialize Docker
	err = json.Unmarshal(*objMap["docker"], &t.Docker)
	if err != nil {
		return err
	}

	// Deserialize Steps by type
	var rawSteps []*json.RawMessage
	err = json.Unmarshal(*objMap["steps"], &rawSteps)
	if err != nil {
		return err
	}
	if err == nil {
		for _, rawMessage := range rawSteps {
			s, err := unmarshalStep(*rawMessage)
			if err != nil {
				return err
			}
			if err == nil {
				err = json.Unmarshal(*rawMessage, s)
				if err != nil {
					return err
				}
				if err == nil {
					t.Steps = append(t.Steps, s)
				}
			}
		}
	}

	// Deserialize ID
	err = json.Unmarshal(*objMap["status"], &t.Status)
	if err != nil {
		return err
	}

	// Deserialize StartedAt
	if objMap["startedAt"] != nil {
		err = json.Unmarshal(*objMap["startedAt"], t.StartedAt)
		if err != nil {
			return err
		}
	}

	// Deserialize UpdatedAt
	if objMap["updatedAt"] != nil {
		err = json.Unmarshal(*objMap["updatedAt"], t.UpdatedAt)
		if err != nil {
			return err
		}
	}

	// Deserialize CompletedAt
	if objMap["completedAt"] != nil {
		err = json.Unmarshal(*objMap["completedAt"], t.CompletedAt)
		if err != nil {
			return err
		}
	}

	// Deserialize ID
	err = json.Unmarshal(*objMap["id"], &t.ID)
	if err != nil {
		return err
	}

	// Deserialize IgnoreErrors
	err = json.Unmarshal(*objMap["ignoreErrors"], &t.IgnoreErrors)
	if err != nil {
		return err
	}

	return nil
}

func (t *Task) Execute(emitter Emitter) error {
	taskWriter := emitter.GetTaskWriter(t)
	defer taskWriter.Close()
	fmt.Fprintf(taskWriter, output.ColorFmt(output.ANSIInfo, "-> running task %s (%s)", "\n"), t.Blueprint.Name, t.ID)
	totalSteps := len(t.Steps)
	for i, step := range t.Steps {
		err := t.executeStep(i+1, totalSteps, emitter, step)
		if err != nil { // TODO: add support for ignoring errors from specific steps in Blueprint
			fmt.Fprintf(taskWriter, output.ColorFmt(output.ANSIError, "-> error in task %s (%s)", "\n"), t.Blueprint.Name, t.ID)
			taskWriter.SetStatus(StateFailed)
			return err
		}
	}
	taskWriter.SetStatus(StateSuccess)
	fmt.Fprintf(taskWriter, output.ColorFmt(output.ANSISuccess, "-> successfully completed task %s (%s)", "\n"), t.Blueprint.Name, t.ID)
	return nil
}

func (t *Task) executeStep(i, totalSteps int, emitter Emitter, step Step) error {
	stepWriter := emitter.GetStepWriter(step)
	defer stepWriter.Close()
	fmt.Fprintf(stepWriter, output.ColorFmt(output.ANSIInfo, "-> running step %d/%d: %s %s (%s)", "\n"), i, totalSteps, step.GetType(), step.GetDescription(), step.GetID())
	step.SetParams(t.parameters)
	err := step.Execute(emitter, t)
	if err != nil {
		stepWriter.SetStatus(StateFailed)
		fmt.Fprintf(stepWriter, output.ColorFmt(output.ANSIError, "-> error in step %s", "\n"), step.GetID())
		return err
	}
	stepWriter.SetStatus(StateSuccess)
	fmt.Fprintf(stepWriter, output.ColorFmt(output.ANSISuccess, "-> successfully completed step %d/%d %s %s (%s)", "\n"), i, totalSteps, step.GetType(), step.GetDescription(), step.GetID())
	return nil
}

func NewTask(
	c *config.Blueprint,
	paramResolver BackupResolver,
	repository *git.Repository,
	branch string,
	commitSha string,
	projectRoot string,
) *Task {
	steps := []Step{
		NewStepSetup(paramResolver, repository, branch, commitSha),
	}
	for _, configStep := range c.Steps {
		switch x := configStep.(type) {
		case *config.StepDockerRun:
			steps = append(steps, NewStepDockerRun(x))
			break
		case *config.StepDockerBuild:
			steps = append(steps, NewStepDockerBuild(x))
			break
		case *config.StepDockerCompose:
			steps = append(steps, NewStepDockerCompose(x, projectRoot))
			break
		case *config.StepDockerPush:
			steps = append(steps, NewStepDockerPush(x))
		}
	}

	return &Task{
		ID:          uuid.NewV4().String(),
		Blueprint:   *c,
		ProjectRoot: projectRoot,
		Steps:       steps,
		parameters:  map[string]*Parameter{},
		Status:      StateWaiting,
	}
}

func (t *Task) UpdateSetup(
	backupResolver BackupResolver,
	repository *git.Repository,
	branch string,
	commitSha string,
) {
	t.Steps[0].(*Setup).backupResolver = backupResolver
	t.Steps[0].(*Setup).repository = repository
	t.Steps[0].(*Setup).branch = branch
	t.Steps[0].(*Setup).commitHash = commitSha
}
