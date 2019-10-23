package builder

import (
	"context"
	"fmt"
)

// BearerAuth implements grpc/credentials.PerRPCCredentials to provide authentication for builder GRPC clients.
type BearerAuth struct {
	token string
}

// NewBearerAuth returns a GRPC mechanism for a static token bearer auth.
func NewBearerAuth(token string) *BearerAuth {
	return &BearerAuth{
		token: token,
	}
}

// GetRequestMetadata returns the additional request metdata for bearer authentication
func (t BearerAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", t.token),
	}, nil
}

// RequireTransportSecurity returns whether or not the authentication mechanism requires transport security.
func (BearerAuth) RequireTransportSecurity() bool {
	return false
}
