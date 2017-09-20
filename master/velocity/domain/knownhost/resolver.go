package knownhost

import (
	"encoding/json"
	"io"

	ut "github.com/go-playground/universal-translator"
	validator "gopkg.in/go-playground/validator.v9"

	"github.com/velocity-ci/velocity/master/velocity/domain"
)

func FromRequest(b io.Reader, validate *validator.Validate, trans ut.Translator) (*domain.KnownHost, error) {
	reqKnownHost := domain.RequestKnownHost{}

	err := json.NewDecoder(b).Decode(&reqKnownHost)
	if err != nil {
		return nil, err
	}

	validate.RegisterValidation("knownHost", ValidateKnownHost)
	validate.RegisterTranslation("knownHost", trans, registerFunc, translationFunc)

	err = validate.Struct(reqKnownHost)

	if err != nil {
		return nil, err
	}

	return reqKnownHost.ToKnownHost(), nil
}
