package auth

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/velocity-ci/velocity/backend/api/domain/user"
)

// FromRequest - Validates and Transforms raw request data into a User struct.
func FromRequest(b io.ReadCloser) (*user.RequestUser, error) {
	requestUser := &user.RequestUser{}

	err := json.NewDecoder(b).Decode(requestUser)
	if err != nil {
		return nil, err
	}

	if len(requestUser.Username) < 3 {
		return nil, errors.New("Invalid credentials")
	}

	if len(requestUser.Password) < 3 {
		return nil, errors.New("Invalid credentials")
	}

	return requestUser, nil
}
