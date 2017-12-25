package project

import (
	"log"
	"os"
	"reflect"

	"github.com/gosimple/slug"
	"github.com/velocity-ci/velocity/backend/api/middleware"
	"github.com/velocity-ci/velocity/backend/velocity"

	ut "github.com/go-playground/universal-translator"
	validator "gopkg.in/go-playground/validator.v9"
)

type Validator struct {
	validator      *validator.Validate
	translator     ut.Translator
	projectManager *Manager
}

func NewValidator(
	validator *validator.Validate,
	trans ut.Translator,
	projectManager *Manager,
) *Validator {
	v := &Validator{
		validator:      validator,
		translator:     trans,
		projectManager: projectManager,
	}
	validator.RegisterValidation("projectUnique", v.ValidateProjectUnique)
	validator.RegisterTranslation("projectUnique", trans, registerFuncUnique, translationFuncUnique)

	validator.RegisterStructValidation(v.ValidateProjectRepository, RequestProject{})
	validator.RegisterTranslation("repository", trans, registerFuncRepository, translationFuncRepository)
	validator.RegisterTranslation("key", trans, registerFuncKey, translationFuncKey)

	return v
}

func (v *Validator) Validate(reqProject *RequestProject) error {
	err := v.validator.Struct(reqProject)
	if _, ok := err.(validator.ValidationErrors); ok {
		return middleware.NewValidationErrors(err.(validator.ValidationErrors), v.translator)
	}
	return err
}

func (v *Validator) ValidateProjectUnique(fl validator.FieldLevel) bool {

	if fl.Field().Type().Name() != "string" {
		return false
	}

	projectName := fl.Field().String()

	_, err := v.projectManager.GetByID(slug.Make(projectName))

	if err != nil {
		return true
	}

	return false
}

func registerFuncUnique(ut ut.Translator) error {
	return ut.Add("projectUnique", "{0} already exists!", true)
}

func translationFuncUnique(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("projectUnique", fe.Field())

	return t
}

func (v *Validator) ValidateProjectRepository(sl validator.StructLevel) {
	p := sl.Current().Interface().(RequestProject)

	apiProject := NewProject(p.Name, velocity.GitRepository{
		Address:    p.Repository,
		PrivateKey: p.PrivateKey,
	})

	_, dir, err := v.projectManager.Sync(&apiProject.Repository, true, false, true, velocity.NewBlankEmitter().NewStreamWriter("clone"))

	if err != nil {
		log.Println(err, reflect.TypeOf(err))
		if _, ok := err.(velocity.SSHKeyError); ok {
			sl.ReportError(p.PrivateKey, "key", "key", "key", "")
		}
		sl.ReportError(p.Repository, "repository", "repository", "repository", "")
	}
	os.RemoveAll(dir)
}

func registerFuncRepository(ut ut.Translator) error {
	return ut.Add("repository", "Could not clone repository! Have you added the host to known hosts?", true)
}

func translationFuncRepository(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("repository", fe.Field())

	return t
}

func registerFuncKey(ut ut.Translator) error {
	return ut.Add("key", "Invalid SSH Key", true)
}

func translationFuncKey(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("key", fe.Field())

	return t
}
