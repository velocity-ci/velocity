package auth_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/velocity-ci/velocity/backend/api/auth"
)

func TestValidFromRequest(t *testing.T) {
	requestJSON := "{\"username\": \"Bob\",\"password\": \"foobar1234\"}"
	b := ioutil.NopCloser(strings.NewReader(requestJSON))

	r, err := auth.FromRequest(b)
	assert.Nil(t, err)
	assert.Equal(t, r.Username, "Bob")
	assert.Equal(t, r.Password, "foobar1234")
}

func TestInvalidJSONFromRequest(t *testing.T) {
	requestJSON := "{\"username\": \"Bob\"\"password\": \"foo\"}"
	b := ioutil.NopCloser(strings.NewReader(requestJSON))

	r, err := auth.FromRequest(b)
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestInvalidUsernameFromRequest(t *testing.T) {
	requestJSON := "{\"username\": \"B\",\"password\": \"foobar1234\"}"
	b := ioutil.NopCloser(strings.NewReader(requestJSON))

	r, err := auth.FromRequest(b)
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestInvalidPasswordFromRequest(t *testing.T) {
	requestJSON := "{\"username\": \"Bob\",\"password\": \"foo\"}"
	b := ioutil.NopCloser(strings.NewReader(requestJSON))

	r, err := auth.FromRequest(b)
	assert.NotNil(t, err)
	assert.Nil(t, r)
}
