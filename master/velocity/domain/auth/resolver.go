package auth

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/VJftw/velocity/master/velocity/domain"
)

// FromRequest - Validates and Transforms raw request data into a User struct.
func FromRequest(b io.ReadCloser) (*domain.User, error) {
	user := &domain.User{}

	requestUser := &requestUser{}

	err := json.NewDecoder(b).Decode(requestUser)
	if err != nil {
		return nil, err
	}

	if len(requestUser.Username) < 3 {
		return nil, errors.New("Invalid credentials")
	}

	if len(requestUser.Password) < 8 {
		return nil, errors.New("Invalid credentials")
	}

	user.Username = requestUser.Username
	user.Password = requestUser.Password

	return user, nil
}

type requestUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
