package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func main() {
	version := flag.Bool("v", false, "Show version")
	list := flag.Bool("l", false, "List tasks")

	flag.Parse()

	gitParams := getGitParams()

	if *version {
		fmt.Println("Version")
		os.Exit(0)
	} else if *list {
		// look for task ymls and parse them into memory.
		tasks := getTasksFromDirectory("./tasks/", gitParams)
		// iterate through tasks in memory and list them.
		for _, task := range tasks {
			fmt.Printf("%s: %s (", task.Name, task.Description)
			for paramName, _ := range task.Parameters {
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
		run(os.Args[2], gitParams)
		break
	default:
		run(os.Args[1], gitParams)
		break
	}
}

func getTasksFromDirectory(dir string, gitParams map[string]velocity.Parameter) []velocity.Task {
	tasks := []velocity.Task{}

	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml") {
			taskYml, _ := ioutil.ReadFile(fmt.Sprintf("%s%s", dir, f.Name()))
			task := velocity.ResolveTaskFromYAML(string(taskYml), gitParams)
			tasks = append(tasks, task)
		}
		return nil
	})

	return tasks
}

func run(taskName string, gitParams map[string]velocity.Parameter) {
	tasks := getTasksFromDirectory("./tasks/", gitParams)

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

	// Resolve parameters
	// reader := bufio.NewReader(os.Stdin)
	// resolvedParams := map[string]velocity.Parameter{}
	// for paramName, p := range t.Parameters {
	// 	// get real value for parameter (ask or from env)
	// 	inputText := ""
	// 	for len(strings.TrimSpace(inputText)) < 1 {
	// 		fmt.Printf("Enter a value for %s (default: %s): ", paramName, p.Value)
	// 		inputText, _ = reader.ReadString('\n')
	// 	}
	// 	p.Value = strings.TrimSpace(inputText)
	// 	resolvedParams[paramName] = p
	// 	t.Parameters[paramName] = p
	// }

	emitter := NewEmitter()
	t.Setup(emitter, &ParameterResolver{}, nil, "")

	// emitter.SetTotalSteps(uint64(len(t.Steps)))
	// Run each step unless they fail (optional)
	for i, step := range t.Steps {
		emitter.SetStepNumber(uint64(i))
		if step.GetType() != "clone" {
			err := step.Execute(emitter, map[string]velocity.Parameter{})
			if err != nil {
				break
			}
		}
	}
}

func getGitParams() map[string]velocity.Parameter {
	path, _ := os.Getwd()

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
	branch := ref.Name().Short()
	describe := shortSHA

	tags, _ := r.Tags()
	defer tags.Close()
	var lastTag *object.Tag
	for {
		t, err := tags.Next()
		if err == io.EOF {
			break
		}

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

	fmt.Printf("----\nGIT_SHA: %s\nGIT_SHORT_SHA: %s\nGIT_BRANCH: %s\nGIT_DESCRIBE: %s\n----\n",
		SHA,
		shortSHA,
		branch,
		describe,
	)

	return map[string]velocity.Parameter{
		"GIT_SHA": velocity.Parameter{
			Value: SHA,
		},
		"GIT_SHORT_SHA": velocity.Parameter{
			Value: shortSHA,
		},
		"GIT_BRANCH": velocity.Parameter{
			Value: branch,
		},
		"GIT_DESCRIBE": velocity.Parameter{
			Value: describe,
		},
	}
}
