package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/velocity-ci/velocity/backend/pkg/grpc/architect"
	"github.com/velocity-ci/velocity/backend/pkg/grpc/architect/db"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	_ "github.com/lib/pq"
	"github.com/velocity-ci/velocity/backend/pkg/auth"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"google.golang.org/grpc"
)

func main() {
	flag.Parse()

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGTERM)
	signal.Notify(stop, syscall.SIGINT)

	log.Println("starting listener")
	lis, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(logging.GetLogger()),
			grpc_auth.StreamServerInterceptor(auth.GRPCInterceptor),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(logging.GetLogger()),
			grpc_auth.UnaryServerInterceptor(auth.GRPCInterceptor),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)

	log.Println("connecting to database")
	db, err := db.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	v1.RegisterBuilderServiceServer(grpcServer, architect.NewBuilderServer())
	v1.RegisterProjectServiceServer(grpcServer, architect.NewProjectServer(db))
	// v1.RegisterRepositoryServiceServer(grpcServer, architect.NewRepositoryServer())
	// v1.RegisterBlueprintServiceServer()
	// v1.RegisterPipelineServiceServer()
	// v1.RegisterBuildServiceServer()

	go func() {
		sig := <-stop
		fmt.Printf("\ncaught signal: %+v\n", sig)
		// fmt.Println("Wait for 2 second to finish processing")
		// time.Sleep(2 * time.Second)
		// os.Exit(0)
		grpcServer.Stop()
		db.Close()
	}()
	log.Println("starting grpc server")
	// TODO: determine whether to use TLS
	grpcServer.Serve(lis)
}
