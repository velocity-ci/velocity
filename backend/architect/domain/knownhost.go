package domain

import (
	ut "github.com/go-playground/universal-translator"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
	"gopkg.in/go-playground/validator.v9"
)

type KnownHost struct {
	UUID              string   `json:"id"`
	Entry             string   `json:"entry" validate:"required,knownHostValid"`
	Hosts             []string `json:"hosts"`
	Comment           string   `json:"comment"`
	SHA256Fingerprint string   `json:"sha256"`
	MD5Fingerprint    string   `json:"md5"`
}

func NewKnownHost(entry string) (*KnownHost, validator.ValidationErrors) {

	k := &KnownHost{
		Entry: entry,
	}

	err := validate.Struct(k)
	if err != nil {
		return nil, err.(validator.ValidationErrors)
	}

	_, hosts, pubKey, comment, _, _ := ssh.ParseKnownHosts([]byte(entry))

	k.UUID = uuid.NewV1().String()
	k.Hosts = hosts
	k.Comment = comment

	if pubKey != nil {
		k.SHA256Fingerprint = ssh.FingerprintSHA256(pubKey)
		k.MD5Fingerprint = ssh.FingerprintLegacyMD5(pubKey)
	}

	return k, nil
}

func (k *KnownHost) RegisterCustomValidations(v *validator.Validate, t ut.Translator) {
	v.RegisterValidation("knownHostValid", ValidateKnownHostValid)
	v.RegisterTranslation("knownHostValid", t, registerFuncValid, translationFuncValid)
}

func ValidateKnownHostValid(fl validator.FieldLevel) bool {
	if fl.Field().Type().Name() != "string" {
		return false
	}

	knownHost := fl.Field().String()
	_, _, _, _, _, err := ssh.ParseKnownHosts([]byte(knownHost))

	if err != nil {
		return false
	}

	return true
}

func registerFuncValid(ut ut.Translator) error {
	return ut.Add("knownHostValid", "{0} is not a valid key!", true)
}

func translationFuncValid(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T("knownHostValid", fe.Field())

	return t
}
