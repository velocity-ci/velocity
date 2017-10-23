package knownhost

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/velocity-ci/velocity/backend/api/middleware"
	"golang.org/x/crypto/ssh"
	validator "gopkg.in/go-playground/validator.v9"
)

type Validator struct {
	validator        *validator.Validate
	translator       ut.Translator
	knownHostManager *Manager
}

func NewValidator(
	validator *validator.Validate,
	trans ut.Translator,
	knownHostManager *Manager,
) *Validator {
	v := &Validator{
		validator:        validator,
		translator:       trans,
		knownHostManager: knownHostManager,
	}

	validator.RegisterValidation("knownHostValid", ValidateKnownHostValid)
	validator.RegisterTranslation("knownHostValid", trans, registerFuncValid, translationFuncValid)

	validator.RegisterValidation("knownHostUnique", v.ValidateKnownHostUnique)
	validator.RegisterTranslation("knownHostUnique", trans, registerFuncUnique, translationFuncUnique)

	return v
}

func (v *Validator) Validate(reqKnownHost *RequestKnownHost) error {
	err := v.validator.Struct(reqKnownHost)
	if _, ok := err.(validator.ValidationErrors); ok {
		return middleware.NewValidationErrors(err.(validator.ValidationErrors), v.translator)
	}
	return err
}

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

func (v *Validator) ValidateKnownHostUnique(fl validator.FieldLevel) bool {

	if fl.Field().Type().Name() != "string" {
		return false
	}

	knownHost := fl.Field().String()

	return !v.knownHostManager.Exists(knownHost)
}

func registerFuncUnique(ut ut.Translator) error {
	return ut.Add("knownHostUnique", "{0} already exists!", true)
}

func translationFuncUnique(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("knownHostUnique", fe.Field())

	return t
}
