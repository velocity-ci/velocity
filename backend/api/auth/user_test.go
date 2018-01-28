package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/api/domain/user"
)

func TestUserValidPassword(t *testing.T) {
	u := &user.User{Username: "Bob"}
	u.HashPassword("foobar")

	assert.True(t, u.ValidatePassword("foobar"))
}

func TestUserInvalidPassword(t *testing.T) {
	u := &user.User{Username: "Bob"}
	u.HashPassword("foobar")

	assert.False(t, u.ValidatePassword("barfoo"))
}
