package domain

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/universal-translator"
	validator "gopkg.in/go-playground/validator.v9"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"
)

func NewValidator() (*validator.Validate, ut.Translator) {
	validate := validator.New()
	en := en.New()
	uni := ut.New(en, en)
	trans, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, trans)

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	return validate, trans
}

type ValidationErrors struct {
	ErrorMap map[string][]string
}

func (v ValidationErrors) Error() string {
	b, _ := json.Marshal(v.ErrorMap)
	return string(b)
}

func NewValidationErrors(errs validator.ValidationErrors, translator ut.Translator) *ValidationErrors {

	m := map[string][]string{}

	for _, err := range errs {
		if _, ok := m[err.Field()]; !ok {
			m[err.Field()] = []string{}
		}
		m[err.Field()] = append(m[err.Field()], err.Translate(translator))
	}

	return &ValidationErrors{ErrorMap: m}
}
