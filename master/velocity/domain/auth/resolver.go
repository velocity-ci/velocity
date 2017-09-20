package auth

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/velocity-ci/velocity/master/velocity/domain"
)

// FromRequest - Validates and Transforms raw request data into a User struct.
func FromRequest(b io.ReadCloser) (*domain.RequestUser, error) {
	requestUser := &domain.RequestUser{}

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

	return requestUser, nil
}
