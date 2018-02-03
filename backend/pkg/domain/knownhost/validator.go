package knownhost

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"golang.org/x/crypto/ssh"
	govalidator "gopkg.in/go-playground/validator.v9"
)

type validator struct {
	validate         *govalidator.Validate
	translator       ut.Translator
	knownHostManager *Manager
}

func newValidator(
	validate *govalidator.Validate,
	trans ut.Translator,
	knownHostManager *Manager,
) *validator {
	v := &validator{
		validate:         validate,
		translator:       trans,
		knownHostManager: knownHostManager,
	}

	v.validate.RegisterValidation("knownHostValid", validateKnownHostValid)
	v.validate.RegisterTranslation("knownHostValid", trans, registerFuncValid, translationFuncValid)

	v.validate.RegisterValidation("knownHostUnique", v.validateKnownHostUnique)
	v.validate.RegisterTranslation("knownHostUnique", trans, registerFuncUnique, translationFuncUnique)

	return v
}

func (v *validator) Validate(k *KnownHost) *domain.ValidationErrors {
	err := v.validate.Struct(k)
	if _, ok := err.(govalidator.ValidationErrors); ok {
		return domain.NewValidationErrors(err.(govalidator.ValidationErrors), v.translator)
	}
	return nil
}

func validateKnownHostValid(fl govalidator.FieldLevel) bool {

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

func translationFuncValid(ut ut.Translator, fe govalidator.FieldError) string {
	t, _ := ut.T("knownHostValid", fe.Field())

	return t
}

func (v *validator) validateKnownHostUnique(fl govalidator.FieldLevel) bool {

	if fl.Field().Type().Name() != "string" {
		return false
	}

	knownHost := fl.Field().String()

	return !v.knownHostManager.Exists(knownHost)
}

func registerFuncUnique(ut ut.Translator) error {
	return ut.Add("knownHostUnique", "{0} already exists!", true)
}

func translationFuncUnique(ut ut.Translator, fe govalidator.FieldError) string {
	t, _ := ut.T("knownHostUnique", fe.Field())

	return t
}
