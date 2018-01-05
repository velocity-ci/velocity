package velocity

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gosimple/slug"
	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type Clone struct {
	BaseStep      `yaml:",inline"`
	GitRepository GitRepository `json:"-" yaml:"-"`
	CommitHash    string        `json:"-" yaml:"-"`
	Submodule     bool          `json:"submodule" yaml:"submodule"`
}

func NewClone() *Clone {
	return &Clone{
		Submodule: false,
		BaseStep: BaseStep{
			Type:          "clone",
			OutputStreams: []string{"clone"},
		},
	}
}

func (c Clone) GetDetails() string {
	return fmt.Sprintf("submodule: %v", c.Submodule)
}

func (c *Clone) Execute(emitter Emitter, params map[string]Parameter) error {
	writer := emitter.NewStreamWriter("clone")
	writer.SetStatus(StateRunning)

	writer.Write([]byte(fmt.Sprintf("%s\n## Cloning %s\n\x1b[0m", infoANSI, c.GitRepository.Address)))

	writer.Write([]byte(fmt.Sprintf("Cloning %s", c.GitRepository.Address)))

	repo, dir, err := GitClone(&c.GitRepository, false, true, c.Submodule, writer)
	if err != nil {
		log.Println(err)
		writer.SetStatus(StateFailed)
		writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
		return err
	}
	log.Println("Done.")
	// defer os.RemoveAll(dir)

	w, err := repo.Worktree()
	if err != nil {
		log.Println(err)
		writer.SetStatus(StateFailed)
		writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
		return err
	}
	log.Printf("Checking out %s", c.CommitHash)
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(c.CommitHash),
	})
	if err != nil {
		log.Println(err)
		writer.SetStatus(StateFailed)
		writer.Write([]byte(fmt.Sprintf("%s\n### FAILED: %s \x1b[0m", errorANSI, err)))
		return err
	}
	log.Println("Done.")

	os.Chdir(dir)
	writer.SetStatus(StateSuccess)
	writer.Write([]byte(fmt.Sprintf("%s\n### SUCCESS \x1b[0m", successANSI)))
	return nil
}

func (cdB *Clone) Validate(params map[string]Parameter) error {
	return nil
}

func (c *Clone) SetParams(params map[string]Parameter) error {
	return nil
}

func (c *Clone) SetGitRepositoryAndCommitHash(r GitRepository, hash string) error {
	c.GitRepository = r
	c.CommitHash = hash
	return nil
}

type SSHKeyError string

func (s SSHKeyError) Error() string {
	return string(s)
}

type GitRepository struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

func GitClone(
	r *GitRepository,
	bare bool,
	full bool,
	submodule bool,
	writer io.Writer,
) (*git.Repository, string, error) {
	psuedoRandom := rand.NewSource(time.Now().UnixNano())
	randNumber := rand.New(psuedoRandom)
	dir := fmt.Sprintf("/tmp/velocity-workspace/velocity_%s-%d", slug.Make(r.Address), randNumber.Int63())
	os.RemoveAll(dir)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return nil, "", err
	}

	isGit := r.Address[:3] == "git"

	var auth transport.AuthMethod

	if isGit {
		log.Printf("git repository: %s", r.Address)
		signer, err := ssh.ParsePrivateKey([]byte(r.PrivateKey))
		if err != nil {
			os.RemoveAll(dir)
			return nil, "", SSHKeyError(err.Error())
		}
		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}

	cloneOpts := &git.CloneOptions{
		URL:      r.Address,
		Auth:     auth,
		Progress: writer,
	}

	if !full {
		cloneOpts.Depth = 1
	}

	if submodule {
		cloneOpts.RecurseSubmodules = 5
	}

	repo, err := git.PlainClone(dir, bare, cloneOpts)

	if err != nil {
		os.RemoveAll(dir)
		return nil, "", err
	}

	if !bare {
		w, _ := repo.Worktree()
		status, _ := w.Status()

		log.Println(status.String())
		w.Reset(&git.ResetOptions{
			Mode: git.HardReset,
		})
	}

	return repo, dir, nil
}
