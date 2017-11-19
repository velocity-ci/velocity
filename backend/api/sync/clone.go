package sync

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/velocity"
	"golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

func GitClone(
	p *project.Project,
	bare bool,
	full bool,
	submodule bool,
	emitter velocity.Emitter,
) (*git.Repository, string, error) {
	psuedoRandom := rand.NewSource(time.Now().UnixNano())
	randNumber := rand.New(psuedoRandom)
	dir := fmt.Sprintf("/tmp/velocity-workspace/velocity_%s-%d", p.ID, randNumber.Int63())
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
