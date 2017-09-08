package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/domain/project/task"
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
			for _, parameter := range task.Parameters {
				fmt.Printf(" %s= %s ", parameter.Name, parameter.Value)
			}
			fmt.Println(")")
			for _, step := range task.Steps {
				fmt.Printf("\t%s| %s: %s\n", step.GetType(), step.GetDescription(), step.GetDetails())
			}
		}
		os.Exit(0)
	}

	switch os.Args[1] {
	case "run":
		run(os.Args[2])
		break
	}
}

func getTasksFromDirectory(dir string) []domain.Task {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	tasks := []domain.Task{}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".yml") || strings.HasSuffix(file.Name(), ".yaml") {
			taskYml, _ := ioutil.ReadFile(fmt.Sprintf("%s%s", dir, file.Name()))
			task := task.ResolveTaskFromYAML(string(taskYml))
			tasks = append(tasks, task)
		}
	}

	return tasks
}

func run(taskName string) {
	tasks := getTasksFromDirectory("./tasks/")

	var task *domain.Task
	// find Task requested
	for _, t := range tasks {
		if t.Name == taskName {
			task = &t
			break
		}
	}

	if task == nil {
		panic(fmt.Sprintf("Task %s not found\n%s", taskName, tasks))
	}

	fmt.Printf("Running task: %s (from: %s)\n", task.Name, taskName)

	// Resolve parameters
	for _, p := range task.Parameters {
		// get real value for parameter (ask or from env)

	}
	task.UpdateParams()

	// Run each step unless they fail (optional)
	for _, step := range task.Steps {
		step.Execute()
	}
}
