package velocity

import (
	"fmt"
	"log"
	"os"
	"time"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type Setup struct {
	BaseStep
	task           *Task
	backupResolver BackupResolver
	repository     *GitRepository
	commitHash     string
}

func NewSetup() *Setup {
	return &Setup{
		BaseStep: BaseStep{
			Type:          "setup",
			OutputStreams: []string{"setup"},
		},
	}
}

func (s *Setup) UnmarshalYamlInterface(y map[interface{}]interface{}) error {
	return nil
}

func (s Setup) GetDetails() string {
	return ""
}

func (s *Setup) Init(
	task *Task,
	backupResolver BackupResolver,
	repository *GitRepository,
	commitHash string,
) {
	s.task = task
	s.backupResolver = backupResolver
	s.repository = repository
	s.commitHash = commitHash
}

func (s *Setup) Execute(emitter Emitter, params map[string]Parameter) error {
	s.task.runID = fmt.Sprintf("vci-%s", time.Now().Format("060102150405"))

	writer := emitter.GetStreamWriter("setup")
	writer.SetStatus(StateRunning)

	// Resolve parameters
	parameters := map[string]Parameter{}
	for _, config := range s.task.Parameters {
		writer.Write([]byte(fmt.Sprintf("Resolving parameter %s", config.GetInfo())))
		params, err := config.GetParameters(writer, s.task.runID, s.backupResolver)
		if err != nil {
			writer.SetStatus(StateFailed)
			log.Printf("could not resolve parameter: %v", err)
			return fmt.Errorf("could not resolve %v", err)
		}
		for _, param := range params {
			parameters[param.Name] = param
			writer.Write([]byte(fmt.Sprintf("Added parameter %s", param.Name)))
		}
	}

	// Update params on steps
	for _, s := range s.task.Steps {
		s.SetParams(parameters)
	}

	// Clone repository if necessary
	if s.repository != nil {
		repo, dir, err := GitClone(s.repository, false, true, s.task.Git.Submodule, writer)
		if err != nil {
			log.Println(err)
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
			return err
		}
		w, err := repo.Worktree()
		if err != nil {
			log.Println(err)
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
			return err
		}
		log.Printf("Checking out %s", s.commitHash)
		err = w.Checkout(&git.CheckoutOptions{
			Hash: plumbing.NewHash(s.commitHash),
		})
		if err != nil {
			log.Println(err)
			writer.SetStatus(StateFailed)
			writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
			return err
		}
		os.Chdir(dir)
	}

	writer.SetStatus(StateSuccess)
	writer.Write([]byte(""))
	writer.Write([]byte("Setup success.\n"))

	return nil
}

func (s *Setup) SetParams(params map[string]Parameter) error {
	return nil
}

func (s Setup) Validate(params map[string]Parameter) error {
	return nil
}
