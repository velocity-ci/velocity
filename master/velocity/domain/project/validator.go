package project

import (
	"log"
	"os"

	"golang.org/x/crypto/ssh"

	ut "github.com/go-playground/universal-translator"
	"github.com/velocity-ci/velocity/master/velocity/persisters"
	validator "gopkg.in/go-playground/validator.v9"
)

func ValidateProjectUnique(fl validator.FieldLevel) bool {

	if fl.Field().Type().Name() != "string" {
		return false
	}

	projectName := fl.Field().String()

	m := NewBoltManager(persisters.GetBoltDB())

	_, err := m.FindByID(idFromName(projectName))

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

func ValidatePrivateKey(fl validator.FieldLevel) bool {

	if fl.Field().Type().Name() != "string" {
		return false
	}

	privateKey := fl.Field().String()

	_, err := ssh.ParsePrivateKey([]byte(privateKey))

	if err != nil {
		return false
	}

	return true
}

func registerFuncPrivateKey(ut ut.Translator) error {
	return ut.Add("privateKey", "{0} is not a valid SSH private key!", true)
}

func translationFuncPrivateKey(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("privateKey", fe.Field())

	return t
}

func ValidateProjectRepository(sl validator.StructLevel) {
	project := sl.Current().Interface().(requestProject)

	_, dir, err := clone(project.Name, project.Repository, project.PrivateKey, true)

	if err != nil {
		log.Println(err)
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
