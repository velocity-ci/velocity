package builder

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/api/proto/v1"
	"google.golang.org/grpc"
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

// BreakRoom joins the breakroom where builders wait to receive tasks.
func BreakRoom(address string, token string) error {
	opts := []grpc.DialOption{grpc.WithPerRPCCredentials(NewBearerAuth(token))}
	if Insecure {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := v1.NewBuilderServiceClient(conn)

	stream, err := client.BreakRoom(context.Background())
	if err != nil {
		return err
	}

	waitC := make(chan struct{})

	go hearbeat(stream)
	go receiveTasks(stream, waitC)

	<-waitC
	return nil
}

func hearbeat(stream v1.BuilderService_BreakRoomClient) {
	for i := 0; i < 5; i++ {
		err := stream.Send(&v1.Heartbeat{
			Timestamp: ptypes.TimestampNow(),
		})
		if err != nil {
			log.Println(err)
		}
		time.Sleep(2 * time.Second)
	}
	stream.CloseSend()

}

func receiveTasks(stream v1.BuilderService_BreakRoomClient, waitC chan struct{}) {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			close(waitC)
			return
		}
		if err != nil {
			log.Fatalf("Failed to receive a note : %v", err)
		}
		t, _ := ptypes.Timestamp(in.Timestamp)
		log.Printf("Got message %s", t)
	}
}
