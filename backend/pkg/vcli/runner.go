package vcli

import (
	"fmt"
	"sync"

	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type runner struct {
	run bool
	wg  *sync.WaitGroup
}

func newRunner(wg *sync.WaitGroup) *runner {
	return &runner{
		run: false,
		wg:  wg,
	}
}

func (r *runner) Run(taskName string) {
	r.run = true
	defer r.wg.Done()
	defer func() { r.run = false }()
	tasks, _ := velocity.GetTasksFromCurrentDir()

	var t *velocity.Task
	// find Task requested
	for _, tsk := range tasks {
		if tsk.Name == taskName {
			t = &tsk
			break
		}
	}

	if t == nil {
		fmt.Printf("Task %s not found in:\n%v\n", taskName, tasks)
		return
	}
	fmt.Printf("Running task: %s\n", t.Name)

	emitter := NewEmitter()

	t.Steps = append([]velocity.Step{velocity.NewSetup()}, t.Steps...)

	// Run each step unless they fail (optional)
	for i, step := range t.Steps {
		if !r.run {
			return
		}
		if step.GetType() == "setup" {
			step.(*velocity.Setup).Init(&ParameterResolver{}, nil, "")
		}
		emitter.SetStepNumber(uint64(i))
		err := step.Execute(emitter, t)
		if err != nil {
			fmt.Printf("encountered error: %s", err)
			return
		}
	}
}

func (r *runner) Stop() {
	if r.run {
		fmt.Printf("\n\nFinishing step\n\n")
		r.run = false
	}
}
