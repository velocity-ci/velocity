package middlewares

import (
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/unrolled/render"
	validator "gopkg.in/go-playground/validator.v9"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"
)

type ResponseErrors struct {
	Errors interface{} `json:"errors"`
}

func newValidator() (*validator.Validate, ut.Translator) {
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

var validatorInstance *validator.Validate
var translatorInstance ut.Translator
var once sync.Once

func GetValidator() (*validator.Validate, ut.Translator) {
	once.Do(func() {
		validatorInstance, translatorInstance = newValidator()
	})

	return validatorInstance, translatorInstance
}

func validationErrorsToJSONMap(errs validator.ValidationErrors) map[string][]string {

	m := map[string][]string{}

	_, t := GetValidator()

	for _, err := range errs {
		if _, ok := m[err.Field()]; !ok {
			m[err.Field()] = []string{}
		}
		m[err.Field()] = append(m[err.Field()], err.Translate(t))
	}

	return m
}

func HandleRequestError(err error, w http.ResponseWriter, render *render.Render) {
	if _, ok := err.(validator.ValidationErrors); ok {
		render.JSON(w, http.StatusBadRequest, validationErrorsToJSONMap(err.(validator.ValidationErrors)))
	} else {
		render.JSON(w, http.StatusBadRequest, "Invalid payload.")
	}
}
