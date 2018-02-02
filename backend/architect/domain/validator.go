package domain

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	validator "gopkg.in/go-playground/validator.v9"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"
)

type Validatable interface {
	// Validate() validator.ValidationErrors
	RegisterCustomValidations(*validator.Validate, ut.Translator)
}

var validate *validator.Validate
var translator ut.Translator
var once sync.Once

func init() {
	once.Do(func() {
		validate = validator.New()
		en := en.New()
		uni := ut.New(en, en)
		translator, _ = uni.GetTranslator("en")
		en_translations.RegisterDefaultTranslations(validate, translator)

		for _, i := range []Validatable{&User{}, &KnownHost{}} {
			i.RegisterCustomValidations(validate, translator)
		}

		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

			if name == "-" {
				return ""
			}

			return name
		})
	})
}

func NewErrorMap(errs validator.ValidationErrors) map[string][]string {
	m := map[string][]string{}

	for _, err := range errs {
		if _, ok := m[err.Field()]; !ok {
			m[err.Field()] = []string{}
		}
		m[err.Field()] = append(m[err.Field()], err.Translate(translator))
	}

	return m
}
