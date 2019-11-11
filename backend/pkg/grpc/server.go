package grpc

import (
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"google.golang.org/grpc"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	// grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
)

// NewServer creates a new GRPC Server with default configuration
func NewServer() (*grpc.Server, error) {
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logging.GetLogger()),
			// grpc_auth.StreamServerInterceptor(auth.GRPCInterceptor),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logging.GetLogger()),
			// grpc_auth.UnaryServerInterceptor(auth.GRPCInterceptor),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)

	return grpcServer, nil
}
