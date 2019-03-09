package build

import (
	"github.com/velocity-ci/velocity/backend/pkg/git"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
)

type Task struct {
	Config config.Task
	// Name        string            `json:"name"`
	// Description string            `json:"description"`
	// Docker      TaskDocker        `json:"docker"`
	// Parameters  []ParameterConfig `json:"parameters"`
	// Steps       []Step            `json:"steps"`

	// ParseErrors      []string `json:"parseErrors"`
	// ValidationErrors []string `json:"validationErrors"`

	Steps      []Step                `json:"steps"`
	Parameters map[string]*Parameter `json:"-"` // Never serialise as resolved
	Docker     TaskDocker

	ProjectRoot string `json:"-"`
	RunID       string `json:"-"`
	// ResolvedParameters map[string]Parameter `json:"-"`
}

func (t *Task) Execute(emitter out.Emitter) error {
	for _, step := range t.Steps {
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
	commitSha string,
	projectRoot string,
) *Task {
	steps := []Step{
		NewStepSetup(paramResolver, repository, commitSha),
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
			steps = append(steps, NewStepDockerCompose(x))
			break
		case *config.StepDockerPush:
			steps = append(steps, NewStepDockerPush(x))
		}
	}

	return &Task{
		Config:      *c,
		ProjectRoot: projectRoot,
		RunID:       "gen",
		Steps:       steps,
		Parameters:  map[string]*Parameter{},
	}
}
