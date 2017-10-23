package auth

import (
	"encoding/json"
	"errors"
	"io"
)

// FromRequest - Validates and Transforms raw request data into a User struct.
func FromRequest(b io.ReadCloser) (*RequestUser, error) {
	requestUser := &RequestUser{}

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
