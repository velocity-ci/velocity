package repository

import "gopkg.in/src-d/go-git.v4/plumbing/transport"

func NewHTTPS(address string) Repository {
	return &httpsRepository{
		address: address,
	}
}

type httpsRepository struct {
	address string
}

func (r *httpsRepository) GetAuthMethod() (transport.AuthMethod, error) {
	return nil, nil
}

func (r *httpsRepository) GetAddress() string {
	return r.address
}
