package repository

import (
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

func NewSSH(address string, privateKey string) Repository {
	return &sshRepository{
		address:    address,
		privateKey: privateKey,
	}
}

type sshRepository struct {
	address    string
	privateKey string
}

func (r *sshRepository) GetAuthMethod() (transport.AuthMethod, error) {
	signer, err := ssh.ParsePrivateKey([]byte(r.privateKey))
	if err != nil {
		return nil, sshKeyError(err.Error())
	}
	return &gitssh.PublicKeys{User: "git", Signer: signer}, nil
}

func (r *sshRepository) GetAddress() string {
	return r.address
}

type sshKeyError string

func (s sshKeyError) Error() string {
	return string(s)
}
