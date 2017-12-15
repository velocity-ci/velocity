package knownhost

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
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

func (r *Resolver) FromRequest(b io.Reader) (KnownHost, error) {
	reqKnownHost := RequestKnownHost{}

	err := json.NewDecoder(b).Decode(&reqKnownHost)
	if err != nil {
		return KnownHost{}, err
	}

	reqKnownHost.Entry = strings.TrimSpace(reqKnownHost.Entry)

	err = r.knownHostValidator.Validate(&reqKnownHost)

	if err != nil {
		return KnownHost{}, err
	}

	return NewKnownHost(reqKnownHost.Entry), nil
}

func (r *Resolver) QueryOptsFromRequest(req *http.Request) KnownHostQuery {
	reqQueries := req.URL.Query()

	amount := uint64(15)
	if a, err := strconv.ParseUint(reqQueries.Get("amount"), 10, 64); err == nil {
		amount = a
	}

	page := uint64(1)
	if p, err := strconv.ParseUint(reqQueries.Get("page"), 10, 64); err == nil {
		page = p
	}

	return KnownHostQuery{
		Amount: amount,
		Page:   page,
	}
}
