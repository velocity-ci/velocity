package knownhost

import (
	ut "github.com/go-playground/universal-translator"
	"golang.org/x/crypto/ssh"
	validator "gopkg.in/go-playground/validator.v9"
)

func ValidateKnownHost(fl validator.FieldLevel) bool {

	if fl.Field().Type().Name() != "string" {
		return false
	}

	knownHost := fl.Field().String()
	_, _, _, _, _, err := ssh.ParseKnownHosts([]byte(knownHost))

	if err != nil {
		return false
	}

	m := NewManager()

	return !m.Exists(knownHost)
}

func registerFunc(ut ut.Translator) error {
	return ut.Add("knownHost", "{0} already exists!", true)
}

func translationFunc(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("knownHost", fe.Field())

	return t
}
