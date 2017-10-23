package project

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/ssh"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type SyncManager struct {
	Sync func(p Project, bare bool) (*git.Repository, string, error)
}

func NewSyncManager(cloneFunc func(p Project, bare bool) (*git.Repository, string, error)) *SyncManager {
	return &SyncManager{
		Sync: cloneFunc,
	}
}

func Clone(p Project, bare bool) (*git.Repository, string, error) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("velocity_%s", p.ID))
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

	repo, err := git.PlainClone(dir, bare, &git.CloneOptions{
		URL:   p.Repository.Address,
		Depth: 1,
		Auth:  auth,
	})

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
