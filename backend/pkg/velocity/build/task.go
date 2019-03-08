package build

import (
	"github.com/velocity-ci/velocity/backend/pkg/git"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
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

func NewTask(
	c *config.Task,
	paramResolver BackupResolver,
	repository *git.Repository,
	commitSha string,
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
	// parameters := map[string]Parameter{}
	// for _, configParameter := range c.Parameters {
	// 	// TODO: resolve, add flag to not resolve
	// 	switch x := configParameter.(type) {
	// 	case *config.ParameterBasic:
	// 		// parameters[x.Name] = x
	// 		break
	// 	case *config.ParameterDerived:
	// 		// resolve (add flag to not resolve)
	// 		break
	// 	}
	// }

	return &Task{
		Config:      *c,
		ProjectRoot: "root",
		RunID:       "gen",
	}
}
