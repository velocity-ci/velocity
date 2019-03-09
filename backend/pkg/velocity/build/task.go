package build

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/git"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
)

type Task struct {
	parameters map[string]*Parameter

	Config config.Task `json:"config"`
	Docker TaskDocker  `json:"docker"`
	Steps  []Step      `json:"steps"`

	StartedAt   *time.Time `json:"startedAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
	CompletedAt *time.Time `json:"completedAt"`

	ProjectRoot string `json:"-"`
	RunID       string `json:"-"`
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
		Config:      *c,
		ProjectRoot: projectRoot,
		RunID:       uuid.NewV4().String(),
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
