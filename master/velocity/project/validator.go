package project

import (
	"log"
	"os"
	"reflect"

	ut "github.com/go-playground/universal-translator"
	"github.com/velocity-ci/velocity/master/velocity/project/git/repository"
	validator "gopkg.in/go-playground/validator.v9"
)

type Validator struct {
	projectManager *BoltManager
}

func (v *Validator) ValidateProjectUnique(fl validator.FieldLevel) bool {

	if fl.Field().Type().Name() != "string" {
		return false
	}

	projectName := fl.Field().String()

	_, err := v.projectManager.FindByID(idFromName(projectName))

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

func ValidateProjectRepository(sl validator.StructLevel) {
	project := sl.Current().Interface().(repository.Repository)

	_, dir, err := Clone(project, true) // use struct to replace func in tests

	if err != nil {
		log.Println(err, reflect.TypeOf(err))
		if _, ok := err.(SSHKeyError); ok {
			sl.ReportError(project.PrivateKey, "key", "key", "key", "")
		}
		sl.ReportError(project.Repository, "repository", "repository", "repository", "")
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
