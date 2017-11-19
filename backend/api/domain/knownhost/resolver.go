package knownhost

import (
	"encoding/json"
	"io"
	"strings"
)

func NewResolver(knownHostValidator *Validator) *Resolver {
	return &Resolver{
		knownHostValidator: knownHostValidator,
	}
}

type Resolver struct {
	knownHostValidator *Validator
}

func (r *Resolver) FromRequest(b io.Reader) (*KnownHost, error) {
	reqKnownHost := RequestKnownHost{}

	err := json.NewDecoder(b).Decode(&reqKnownHost)
	if err != nil {
		return nil, err
	}

	reqKnownHost.Entry = strings.TrimSpace(reqKnownHost.Entry)

	err = r.knownHostValidator.Validate(&reqKnownHost)

	if err != nil {
		return nil, err
	}

	return &KnownHost{
		Entry: reqKnownHost.Entry,
	}, nil
}
