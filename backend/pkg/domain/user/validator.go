package user

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	govalidator "gopkg.in/go-playground/validator.v9"
)

type validator struct {
	validate    *govalidator.Validate
	translator  ut.Translator
	userManager *Manager
}

func newValidator(
	validate *govalidator.Validate,
	trans ut.Translator,
	userManager *Manager,
) *validator {
	v := &validator{
		validate:    validate,
		translator:  trans,
		userManager: userManager,
	}

	v.validate.RegisterValidation("userUnique", v.validateUserUnique)
	v.validate.RegisterTranslation("userUnique", trans, registerFuncUnique, translationFuncUnique)

	return v
}

func (v *validator) Validate(u *User) *domain.ValidationErrors {
	err := v.validate.Struct(u)
	if _, ok := err.(govalidator.ValidationErrors); ok {
		return domain.NewValidationErrors(err.(govalidator.ValidationErrors), v.translator)
	}
	return nil
}

func (v *validator) validateUserUnique(fl govalidator.FieldLevel) bool {

	if fl.Field().Type().Name() != "string" {
		return false
	}

	username := fl.Field().String()
	if _, err := v.userManager.GetByUsername(username); err != nil {
		return true
	}

	return false
}

func registerFuncUnique(ut ut.Translator) error {
	return ut.Add("userUnique", "{0} already exists!", true)
}

func translationFuncUnique(ut ut.Translator, fe govalidator.FieldError) string {
	t, _ := ut.T("userUnique", fe.Field())

	return t
}
