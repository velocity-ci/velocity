package project

import (
	"os"

	ut "github.com/go-playground/universal-translator"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	govalidator "gopkg.in/go-playground/validator.v9"
)

type validator struct {
	validate       *govalidator.Validate
	translator     ut.Translator
	projectManager *Manager
}

func newValidator(
	validate *govalidator.Validate,
	trans ut.Translator,
	projectManager *Manager,
) *validator {
	v := &validator{
		validate:       validate,
		translator:     trans,
		projectManager: projectManager,
	}

	v.validate.RegisterValidation("projectUnique", v.validateProjectUnique)
	v.validate.RegisterTranslation("projectUnique", trans, registerFuncUnique, translationFuncUnique)

	v.validate.RegisterStructValidation(v.validateProjectRepository, Project{})
	v.validate.RegisterTranslation("gitRepository", trans, registerFuncRepository, translationFuncRepository)
	v.validate.RegisterTranslation("sshPrivateKey", trans, registerFuncKey, translationFuncKey)

	return v
}

func (v *validator) Validate(p *Project) *domain.ValidationErrors {
	err := v.validate.Struct(p)
	if _, ok := err.(govalidator.ValidationErrors); ok {
		return domain.NewValidationErrors(err.(govalidator.ValidationErrors), v.translator)
	}
	return nil
}

func (v *validator) validateProjectUnique(fl govalidator.FieldLevel) bool {

	if fl.Field().Type().Name() != "string" {
		return false
	}

	projectName := fl.Field().String()
	if _, err := v.projectManager.GetByName(projectName); err != nil {
		return true
	}

	return false
}

func registerFuncUnique(ut ut.Translator) error {
	return ut.Add("projectUnique", "{0} already exists!", true)
}

func translationFuncUnique(ut ut.Translator, fe govalidator.FieldError) string {
	t, _ := ut.T("projectUnique", fe.Field())

	return t
}

func (v *validator) validateProjectRepository(sl govalidator.StructLevel) {
	p := sl.Current().Interface().(Project)

	if len(p.Config.Address) < 4 {
		sl.ReportError(p.Config.Address, "repository", "repository", "repository", "")
		return
	}

	repo, err := v.projectManager.clone(&p.Config, true, false, true, velocity.NewBlankEmitter().GetStreamWriter("clone"))

	if err != nil {
		if _, ok := err.(velocity.SSHKeyError); ok {
			sl.ReportError(p.Config.PrivateKey, "key", "key", "sshPrivateKey", err.Error())
		} else {
			sl.ReportError(p.Config.Address, "repository", "repository", "gitRepository", "")
		}
	}
	os.RemoveAll(repo.Directory)
}

func registerFuncRepository(ut ut.Translator) error {
	return ut.Add("gitRepository", "Could not clone repository! Have you added the host to known hosts?", true)
}

func translationFuncRepository(ut ut.Translator, fe govalidator.FieldError) string {
	t, _ := ut.T("gitRepository", fe.Field())

	return t
}

func registerFuncKey(ut ut.Translator) error {
	return ut.Add("sshPrivateKey", "Invalid SSH Key: {0}", true)
}

func translationFuncKey(ut ut.Translator, fe govalidator.FieldError) string {
	t, _ := ut.T("sshPrivateKey", fe.Field())

	return t
}
