package middleware

import (
	"encoding/json"
	"net/http"

	ut "github.com/go-playground/universal-translator"
	"github.com/unrolled/render"
	validator "gopkg.in/go-playground/validator.v9"
)

func HandleRequestError(err error, w http.ResponseWriter, r *render.Render) {
	if _, ok := err.(ValidationErrors); ok {
		r.JSON(w, http.StatusBadRequest, err.(ValidationErrors).errorMap)
	} else {
		r.JSON(w, http.StatusBadRequest, "Invalid payload.")
	}
}

type ValidationErrors struct {
	errorMap map[string][]string
}

func (v ValidationErrors) Error() string {
	b, _ := json.Marshal(v.errorMap)
	return string(b)
}

func NewValidationErrors(errs validator.ValidationErrors, translator ut.Translator) ValidationErrors {

	m := map[string][]string{}

	for _, err := range errs {
		if _, ok := m[err.Field()]; !ok {
			m[err.Field()] = []string{}
		}
		m[err.Field()] = append(m[err.Field()], err.Translate(translator))
	}

	return ValidationErrors{errorMap: m}
}
