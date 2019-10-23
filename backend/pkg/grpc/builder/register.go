package builder

import (
	"context"

	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
)

// Register registers a new builder and returns the JWT to authenticate with.
func Register(address string) (string, error) {
	conn, err := NewConn(address)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	client := v1.NewBuilderServiceClient(conn)

	reg, err := client.Register(context.Background(), &v1.RegisterRequest{})
	if err != nil {
		return "", err
	}
	token := reg.GetToken()

	return token, nil
}
