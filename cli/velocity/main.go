package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/domain/task"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func main() {
	version := flag.Bool("v", false, "Show version")
	list := flag.Bool("l", false, "List tasks")

	flag.Parse()

	getGitParams()

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
	tasks := []domain.Task{}

	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml") {
			taskYml, _ := ioutil.ReadFile(fmt.Sprintf("%s%s", dir, f.Name()))
			task := task.ResolveTaskFromYAML(string(taskYml), getGitParams())
			tasks = append(tasks, task)
		}
		return nil
	})

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
	reader := bufio.NewReader(os.Stdin)
	resolvedParams := []domain.Parameter{}
	for _, p := range task.Parameters {
		// get real value for parameter (ask or from env)
		inputText := ""
		for len(strings.TrimSpace(inputText)) < 1 {
			fmt.Printf("Enter a value for %s (default: %s): ", p.Name, p.Value)
			inputText, _ = reader.ReadString('\n')
		}
		p.Value = strings.TrimSpace(inputText)
		resolvedParams = append(resolvedParams, p)
	}

	gitParams := getGitParams()

	task.Parameters = append(resolvedParams, gitParams...)
	task.UpdateParams()
	task.SetEmitter(func(s string) { fmt.Printf("    %s\n", s) })

	// Run each step unless they fail (optional)
	for _, step := range task.Steps {
		step.Execute()
	}
}

func getGitParams() []domain.Parameter {
	path, _ := os.Getwd()
	if os.Getenv("SIB_CWD") != "" {
		path = os.Getenv("SIB_CWD")
	}

	// We instance a new repository targeting the given path (the .git folder)
	r, err := git.PlainOpen(fmt.Sprintf("%s/", path))
	if err != nil {
		panic(err)
	}

	// ... retrieving the HEAD reference
	ref, err := r.Head()
	if err != nil {
		panic(err)
	}
	SHA := ref.Hash().String()
	shortSHA := SHA[:7]
	fmt.Println(SHA)
	fmt.Println(shortSHA)
	branch := ref.Name().Short()
	fmt.Println(branch)

	describe := shortSHA

	tags, _ := r.Tags()
	defer tags.Close()
	var lastTag *object.Tag
	for {
		t, err := tags.Next()
		if err == io.EOF {
			break
		}

		fmt.Println(t)

		tObj, err := r.TagObject(t.Hash())
		if err != nil {
			panic(err)
		}

		c, _ := tObj.Commit()
		if c.Hash.String() == SHA {
			describe = tObj.Name
		}
		lastTag = tObj
	}

	if describe == shortSHA {
		if lastTag == nil {
			describe = shortSHA
		} else {
			describe = fmt.Sprintf("%s+%s", lastTag.Name, shortSHA)
		}
	}

	fmt.Println(describe)

	return []domain.Parameter{
		domain.Parameter{
			Name:  "GIT_SHA",
			Value: SHA,
		},
		domain.Parameter{
			Name:  "GIT_SHORT_SHA",
			Value: shortSHA,
		},
		domain.Parameter{
			Name:  "GIT_BRANCH",
			Value: branch,
		},
		domain.Parameter{
			Name:  "GIT_DESCRIBE",
			Value: describe,
		},
	}
}
