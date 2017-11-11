package velocity

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type Clone struct {
	BaseStep  `yaml:",inline"`
	Build     *Build
	Submodule bool `json:"submodule" yaml:"submodule"`
}

func NewClone() *Clone {
	return &Clone{
		Submodule: false,
	}
}

func (c Clone) GetType() string {
	return "clone"
}

func (c Clone) GetDescription() string {
	return c.Description
}

func (c Clone) GetDetails() string {
	return fmt.Sprintf("submodule: %v", c.Submodule)
}

func (c *Clone) Execute(emitter Emitter, params map[string]Parameter) error {
	emitter.Write([]byte(fmt.Sprintf("%s\n## %s\n\x1b[0m", infoANSI, c.Description)))

	log.Printf("Cloning %s", c.Build.Project.Repository.Address)

	repo, dir, err := GitClone(c.Build.Project, false, true, c.Submodule, emitter)
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Println("Done.")
	// defer os.RemoveAll(dir)

	w, err := repo.Worktree()
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Printf("Checking out %s", c.Build.CommitHash)
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(c.Build.CommitHash),
	})
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Println("Done.")

	os.Chdir(dir)
	return nil
}

func (cdB *Clone) Validate(params map[string]Parameter) error {
	return nil
}

func (c *Clone) SetParams(params map[string]Parameter) error {
	return nil
}

func (c *Clone) SetBuild(b *Build) error {
	c.Build = b
	return nil
}

func GitClone(
	p *Project,
	bare bool,
	full bool,
	submodule bool,
	emitter Emitter,
) (*git.Repository, string, error) {
	psuedoRandom := rand.NewSource(time.Now().UnixNano())
	randNumber := rand.New(psuedoRandom)
	dir := fmt.Sprintf("/var/velocity-workspace/velocity_%s-%d", p.ID, randNumber.Int63())
	os.RemoveAll(dir)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return nil, "", err
	}

	isGit := p.Repository.Address[:3] == "git"

	var auth transport.AuthMethod

	if isGit {
		log.Printf("git repository: %s", p.Repository.Address)
		signer, err := ssh.ParsePrivateKey([]byte(p.Repository.PrivateKey))
		if err != nil {
			os.RemoveAll(dir)
			return nil, "", SSHKeyError(err.Error())
		}
		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}

	cloneOpts := &git.CloneOptions{
		URL:      p.Repository.Address,
		Auth:     auth,
		Progress: emitter,
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

	return repo, dir, nil

}

type SSHKeyError string

func (s SSHKeyError) Error() string {
	return string(s)
}
