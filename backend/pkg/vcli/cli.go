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
			// fmt.Printf("-> %s has warnings:\n", task.Name)
			for _, warn := range task.ValidationWarnings {
				fmt.Print(colorFmt(ansiWarn, fmt.Sprintf("  %s\n", warn)))
			}
			fmt.Println()
			continue
		} else {
			// fmt.Printf("-> %s\n", task.Name)
		}
		fmt.Printf("  %s\n\n", task.Description)

		// for _, paramName := range task.Parameters {
		// 	fmt.Printf(" %s ", paramName)
		// }
		// fmt.Println(")")
		// for _, step := range task.Steps {
		// 	fmt.Printf("\t%s| %s: %s\n", step.GetType(), step.GetDescription(), step.GetDetails())
		// }
		// fmt.Println()
	}
	return nil
}

// func run(c *cli.Context) error {
// 	c.wg.Add(1)
// 	c.runner.Run(os.Args[2])
// 	return nil
// }

func validate(c *cli.Context) error {
	return nil
}

const (
	ansiSuccess = "\x1b[1m\x1b[49m\x1b[32m"
	ansiError   = "\x1b[1m\x1b[49m\x1b[31m"
	ansiWarn    = "\x1b[1m\x1b[49m\x1b[33m"
	ansiInfo    = "\x1b[1m\x1b[49m\x1b[94m"
)

func colorFmt(ansiColor, format string) string {
	return fmt.Sprintf("%s%s\x1b[0m", ansiColor, format)
}
