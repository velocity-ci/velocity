package knownhost

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/middlewares"
)

type requestKnownHost struct {
	Entry string `json:"entry" validate:"required,knownHostValid,knownHostUnique,min=10"`
}

func FromRequest(b io.Reader) (*domain.KnownHost, error) {
	reqKnownHost := requestKnownHost{}

	err := json.NewDecoder(b).Decode(&reqKnownHost)
	if err != nil {
		return nil, err
	}

	reqKnownHost.Entry = strings.TrimSpace(reqKnownHost.Entry)

	validate, trans := middlewares.GetValidator()

	validate.RegisterValidation("knownHostValid", ValidateKnownHostValid)
	validate.RegisterTranslation("knownHostValid", trans, registerFuncValid, translationFuncValid)

	validate.RegisterValidation("knownHostUnique", ValidateKnownHostUnique)
	validate.RegisterTranslation("knownHostUnique", trans, registerFuncUnique, translationFuncUnique)

	err = validate.Struct(reqKnownHost)

	if err != nil {
		return nil, err
	}

	return &domain.KnownHost{
		Entry: reqKnownHost.Entry,
	}, nil
}
