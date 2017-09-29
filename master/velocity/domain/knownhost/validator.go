package knownhost

import (
	"log"
	"os"

	ut "github.com/go-playground/universal-translator"
	"golang.org/x/crypto/ssh"
	validator "gopkg.in/go-playground/validator.v9"
)

func ValidateKnownHostValid(fl validator.FieldLevel) bool {

	if fl.Field().Type().Name() != "string" {
		return false
	}

	knownHost := fl.Field().String()
	_, _, _, _, _, err := ssh.ParseKnownHosts([]byte(knownHost))

	if err != nil {
		return false
	}

	return true

}

func registerFuncValid(ut ut.Translator) error {
	return ut.Add("knownHostValid", "{0} is not a valid key!", true)
}

func translationFuncValid(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("knownHostValid", fe.Field())

	return t
}

func ValidateKnownHostUnique(fl validator.FieldLevel) bool {

	if fl.Field().Type().Name() != "string" {
		return false
	}

	knownHost := fl.Field().String()

	m := NewManager(log.New(os.Stdout, "[files]", log.Lshortfile))

	return !m.Exists(knownHost)
}

func registerFuncUnique(ut ut.Translator) error {
	return ut.Add("knownHostUnique", "{0} already exists!", true)
}

func translationFuncUnique(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("knownHostUnique", fe.Field())

	return t
}
