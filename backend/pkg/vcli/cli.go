package vcli

import (
	"fmt"
	"os"
	"sync"

	"github.com/urfave/cli"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type CLI struct {
	wg     sync.WaitGroup
	runner *runner
}

func New() *CLI {
	c := &CLI{
		wg: sync.WaitGroup{},
	}
	c.runner = newRunner(&c.wg)
	return c
}

func (c *CLI) Start(quit chan os.Signal) {
	// c.routeFlags()

	quit <- os.Interrupt
}

func (c *CLI) Stop() error {
	c.runner.Stop()
	c.wg.Wait()
	return nil
}

func (c *CLI) Run() {

	switch os.Args[1] {
	case "run":
		c.wg.Add(1)
		c.runner.Run(os.Args[2])
		break
	default:
		c.wg.Add(1)
		c.runner.Run(os.Args[1])
		break
	}

	c.Stop()
}

func List(c *cli.Context) error {
	tasks, err := velocity.GetTasksFromCurrentDir()
	if err != nil {
		return err
	}
	// iterate through tasks in memory and list them.
	for _, task := range tasks {
		fmt.Print(colorFmt(ansiInfo, fmt.Sprintf("-> %s", task.Name)))
		if len(task.ValidationErrors) > 0 {
			fmt.Printf(" has errors:\n")
			for _, err := range task.ValidationErrors {
				fmt.Print(colorFmt(ansiError, fmt.Sprintf("  %s\n", err)))
			}
			fmt.Println()
			continue
		}
		if len(task.ValidationWarnings) > 0 {
			fmt.Printf(" has warnings:\n")
			for _, warn := range task.ValidationWarnings {
				fmt.Print(colorFmt(ansiWarn, fmt.Sprintf("  %s\n", warn)))
			}
			fmt.Println()
			continue
		}
		fmt.Printf("  %s\n\n", task.Description)

	}
	return nil
}

func Info(c *cli.Context) error {
	basicParams := velocity.GetBasicParams()
	for key, val := range basicParams {
		fmt.Printf("%s: %s\n", key, val.Value)
	}
	return nil
}

func Run(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("incorrect amount of args")
	}
	tasks, err := velocity.GetTasksFromCurrentDir()
	if err != nil {
		return err
	}
	task, err := getRequestedTaskByName(c.Args().Get(0), tasks)
	if err != nil {
		return err
	}
	fmt.Print(colorFmt(ansiInfo, fmt.Sprintf("-> running: %s\n", task.Name)))
	emitter := NewEmitter()

	task.Steps = append([]velocity.Step{velocity.NewSetup()}, task.Steps...)
	for i, step := range task.Steps {
		// if !r.run {
		// 	return
		// }
		if step.GetType() == "setup" {
			step.(*velocity.Setup).Init(&ParameterResolver{}, nil, "")
		}
		emitter.SetStepNumber(uint64(i))
		step.SetProjectRoot(task.ProjectRoot)
		err := step.Execute(emitter, task)
		if err != nil {
			fmt.Printf("encountered error: %s", err)
			return err
		}
	}

	return nil
}

func getRequestedTaskByName(taskName string, tasks []velocity.Task) (*velocity.Task, error) {
	for _, t := range tasks {
		if t.Name == taskName {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("could not find %s", taskName)
}

// func run(c *cli.Context) error {
// 	c.wg.Add(1)
// 	c.runner.Run(os.Args[2])
// 	return nil
// }

const (
	ansiSuccess = "\x1b[1m\x1b[49m\x1b[32m"
	ansiError   = "\x1b[1m\x1b[49m\x1b[31m"
	ansiWarn    = "\x1b[1m\x1b[49m\x1b[33m"
	ansiInfo    = "\x1b[1m\x1b[49m\x1b[94m"
)

func colorFmt(ansiColor, format string) string {
	return fmt.Sprintf("%s%s\x1b[0m", ansiColor, format)
}
