package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/api/auth"
)

func TestGenerateRandomString(t *testing.T) {
	a := auth.GenerateRandomString(8)
	b := auth.GenerateRandomString(8)

	assert.NotEqual(t, a, b)
}
