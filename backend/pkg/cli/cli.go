package cli

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	yaml "gopkg.in/yaml.v2"
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
	c.routeFlags()

	quit <- os.Interrupt
}

func (c *CLI) Stop() error {
	c.runner.Stop()
	c.wg.Wait()
	return nil
}

func (c *CLI) routeFlags() {
	version := flag.Bool("v", false, "Show version")
	list := flag.Bool("l", false, "List tasks")

	flag.Parse()

	if *version {
		fmt.Printf("Version: %s\n", "alpha")
		os.Exit(0)
	}

	if *list {
		tasks := getTasksFromDirectory("./tasks/")
		// iterate through tasks in memory and list them.
		for _, task := range tasks {
			fmt.Printf("%s: %s (", task.Name, task.Description)
			for _, paramName := range task.Parameters {
				fmt.Printf(" %s ", paramName)
			}
			fmt.Println(")")
			for _, step := range task.Steps {
				fmt.Printf("\t%s| %s: %s\n", step.GetType(), step.GetDescription(), step.GetDetails())
			}
			fmt.Println()
		}
		os.Exit(0)
	}

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

func getTasksFromDirectory(dir string) []velocity.Task {
	tasks := []velocity.Task{}

	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml") {
			taskYml, _ := ioutil.ReadFile(fmt.Sprintf("%s%s", dir, f.Name()))
			var t velocity.Task
			err := yaml.Unmarshal(taskYml, &t)
			if err != nil {
				log.Println(err)
			} else {
				tasks = append(tasks, t)
			}
		}
		return nil
	})

	return tasks
}
