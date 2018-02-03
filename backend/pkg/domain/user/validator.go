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

	return v
}

func (v *validator) Validate(u *User) *domain.ValidationErrors {
	err := v.validate.Struct(u)
	if _, ok := err.(govalidator.ValidationErrors); ok {
		return domain.NewValidationErrors(err.(govalidator.ValidationErrors), v.translator)
	}
	return nil
}
