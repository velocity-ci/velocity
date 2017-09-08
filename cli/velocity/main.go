package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/VJftw/velocity/master/velocity/domain"
	"github.com/VJftw/velocity/master/velocity/domain/project/task"
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
				fmt.Printf(" %s= %s ", parameter.Name, parameter.Default)
			}
			fmt.Println(")")
			for _, step := range task.Steps {
				fmt.Printf("\t%s| %s\n", step.GetType(), step.GetDescription())
			}
		}
		os.Exit(0)
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
