package build

import (
	"encoding/json"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/git"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
)

type Task struct {
	ID         string `json:"id"`
	parameters map[string]*Parameter

	Config config.Task `json:"config"`
	Docker TaskDocker  `json:"docker"`
	Steps  []Step      `json:"steps"`

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

	// Deserialize Config
	err = json.Unmarshal(*objMap["config"], &t.Config)
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

	// Deserialize StartedAt
	err = json.Unmarshal(*objMap["startedAt"], t.StartedAt)
	if err != nil {
		return err
	}

	// Deserialize UpdatedAt
	err = json.Unmarshal(*objMap["updatedAt"], t.UpdatedAt)
	if err != nil {
		return err
	}

	// Deserialize CompletedAt
	err = json.Unmarshal(*objMap["completedAt"], t.CompletedAt)
	if err != nil {
		return err
	}

	// Deserialize ID
	err = json.Unmarshal(*objMap["id"], &t.ID)
	if err != nil {
		return err
	}

	return nil
}

func (t *Task) Execute(emitter out.Emitter) error {
	for _, step := range t.Steps {
		step.SetParams(t.parameters)
		err := step.Execute(emitter, t)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewTask(
	c *config.Task,
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
		Config:      *c,
		ProjectRoot: projectRoot,
		Steps:       steps,
		parameters:  map[string]*Parameter{},
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
