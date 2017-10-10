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

func Clone(name string, repositoryAddress string, key string, bare bool) (*git.Repository, string, error) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("velocity_%s", idFromName(name)))
	if err != nil {
		log.Fatal(err)
		return nil, "", err
	}

	isGit := repositoryAddress[:3] == "git"

	var auth transport.AuthMethod

	if isGit {
		log.Printf("git repository: %s", repositoryAddress)
		signer, err := ssh.ParsePrivateKey([]byte(key))
		if err != nil {
			os.RemoveAll(dir)
			return nil, "", SSHKeyError(err.Error())
		}
		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}

	repo, err := git.PlainClone(dir, bare, &git.CloneOptions{
		URL:   repositoryAddress,
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
