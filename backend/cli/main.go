package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/velocity-ci/velocity/backend/velocity"
)

func main() {
	version := flag.Bool("v", false, "Show version")
	list := flag.Bool("l", false, "List tasks")

	flag.Parse()

	if *version {
		fmt.Println("Version")
		os.Exit(0)
	} else if *list {
		// look for task ymls and parse them into memory.
		tasks := getTasksFromDirectory("./tasks/")
		// iterate through tasks in memory and list them.
		for _, task := range tasks {
			fmt.Printf("%s: %s (", task.Name, task.Description)
			for paramName := range task.Parameters {
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
		run(os.Args[2])
		break
	default:
		run(os.Args[1])
		break
	}
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

func run(taskName string) {
	tasks := getTasksFromDirectory("./tasks/")

	var t *velocity.Task
	// find Task requested
	for _, tsk := range tasks {
		if tsk.Name == taskName {
			t = &tsk
			break
		}
	}

	if t == nil {
		panic(fmt.Sprintf("Task %s not found\n%v", taskName, tasks))
	}

	fmt.Printf("Running task: %s (from: %s)\n", t.Name, taskName)

	emitter := NewEmitter()

	t.Steps = append([]velocity.Step{velocity.NewSetup()}, t.Steps...)

	// Run each step unless they fail (optional)
	for i, step := range t.Steps {
		if step.GetType() == "setup" {
			step.(*velocity.Setup).Init(&ParameterResolver{}, nil, "")
		}
		emitter.SetStepNumber(uint64(i))
		err := step.Execute(emitter, t)
		if err != nil {
			break
		}
	}
}
