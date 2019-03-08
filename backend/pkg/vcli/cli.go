package vcli

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/urfave/cli"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/build"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
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
	tasks, _, err := config.GetTasksFromCurrentDir()
	if err != nil {
		return err
	}

	if !c.Bool("machine-readable") {
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
			if len(task.ValidationErrors) > 0 {
				fmt.Printf(" has warnings:\n")
				for _, warn := range task.ValidationErrors {
					fmt.Print(colorFmt(ansiWarn, fmt.Sprintf("  %s\n", warn)))
				}
				fmt.Println()
				continue
			}
			if len(task.ParseErrors) > 0 {
				fmt.Printf(" has warnings:\n")
				for _, warn := range task.ParseErrors {
					fmt.Print(colorFmt(ansiWarn, fmt.Sprintf("  %s\n", warn)))
				}
				fmt.Println()
				continue
			}
			fmt.Printf("  %s\n\n", task.Description)

		}
	} else {
		jsonBytes, err := json.MarshalIndent(tasks, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", jsonBytes)
	}

	return nil
}

// func RunCompletion(c *cli.Context) {
// 	if c.NArg() > 0 {
// 		return
// 	}
// 	tasks, _ := task.GetTasksFromCurrentDir()
// 	for _, t := range tasks {
// 		fmt.Println(t.Name)
// 	}
// }

func Info(c *cli.Context) error {
	_, projectRoot, err := config.GetTasksFromCurrentDir()
	if err != nil {
		return err
	}
	emitter := NewEmitter()

	basicParams, err := build.GetGlobalParams(emitter.GetStreamWriter("setup"), projectRoot)
	if err != nil {
		return err
	}
	for key, val := range basicParams {
		fmt.Printf("%s: %s\n", key, val.Value)
	}
	return nil
}

func Run(c *cli.Context) error {
	if c.NArg() != 1 {
		return fmt.Errorf("incorrect amount of args")
	}
	// tasks, err := task.GetTasksFromCurrentDir()
	// if err != nil {
	// 	return err
	// }
	// t, err := getRequestedTaskByName(c.Args().Get(0), tasks)
	// if err != nil {
	// 	return err
	// }
	// fmt.Print(colorFmt(ansiInfo, fmt.Sprintf("-> running: %s\n", t.Name)))
	// emitter := NewEmitter()

	// t.Steps = append([]task.Step{task.NewSetup()}, t.Steps...)
	// for i, step := range t.Steps {
	// 	// if !r.run {
	// 	// 	return
	// 	// }
	// 	if step.GetType() == "setup" {
	// 		step.(*task.Setup).Init(&ParameterResolver{}, nil, "")
	// 	}
	// 	emitter.SetStepNumber(uint64(i))
	// 	step.SetProjectRoot(t.ProjectRoot)
	// 	err := step.Execute(emitter, t)
	// 	if err != nil {
	// 		fmt.Printf("encountered error: %s", err)
	// 		return err
	// 	}
	// }

	return nil
}

func getRequestedTaskByName(taskName string, tasks []*config.Task) (*config.Task, error) {
	for _, t := range tasks {
		if t.Name == taskName {
			return t, nil
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
