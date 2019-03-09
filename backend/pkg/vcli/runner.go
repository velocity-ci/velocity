package vcli

import (
	"fmt"
	"sync"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/build"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
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
	tasks, projectRoot, err := config.GetTasksFromCurrentDir()
	if err != nil {
		fmt.Printf("encountered error: %s", err)
		return
	}

	var configTaskToRun *config.Task
	for _, tsk := range tasks {
		if tsk.Name == taskName {
			configTaskToRun = tsk
			break
		}
	}

	if configTaskToRun == nil {
		fmt.Printf("Task %s not found in:\n%v\n", taskName, tasks)
		return
	}
	fmt.Printf("Running task: %s\n", configTaskToRun.Name)

	emitter := NewEmitter()

	task := build.NewTask(
		configTaskToRun,
		&ParameterResolver{},
		nil,
		"",
		projectRoot,
	)

	err = task.Execute(emitter)
	if err != nil {
		fmt.Printf("error: %s", err)
	}

}

func (r *runner) Stop() {
	if r.run {
		fmt.Printf("\n\nFinishing step\n\n")
		r.run = false
	}
}
