package repository

import "gopkg.in/src-d/go-git.v4/plumbing/transport"

type Repository interface {
	GetAuthMethod() (transport.AuthMethod, error)
	GetAddress() string
}
