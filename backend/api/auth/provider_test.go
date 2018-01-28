package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/api/auth"
)

func TestNewToken(t *testing.T) {

	// When generating a new AuthToken
	authToken := auth.NewAuthToken("Bob")

	// Then the username should be correct
	assert.Equal(t, authToken.Username, "Bob")
	// And there should be a token string
	assert.NotEmpty(t, authToken.Token)
	// And the token should expire in 2 days
	assert.WithinDuration(
		t,
		time.Now().Add(time.Hour*24*2),
		authToken.Expires,
		time.Duration(1)*time.Second,
	)
}
