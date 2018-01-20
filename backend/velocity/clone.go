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
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

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

		w.Reset(&git.ResetOptions{
			Mode: git.HardReset,
		})
	}

	return repo, dir, nil
}
