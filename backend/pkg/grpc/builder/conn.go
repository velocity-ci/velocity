package builder

import "google.golang.org/grpc"

var (
	// Insecure represents if the builder should use an insecure GRPC connection.
	Insecure = false
)

// NewConn returns a new GRPC connection.
func NewConn(address string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{}
	if Insecure {
		opts = append(opts, grpc.WithInsecure())
	}
	return grpc.Dial(address, opts...)
}
