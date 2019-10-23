package builder

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/golang/protobuf/ptypes"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
	"google.golang.org/grpc"
)

// BreakRoom joins the breakroom where builders wait to receive tasks.
func BreakRoom(address string, token string, gracefulStop chan os.Signal) error {
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

	go hearbeat(stream, gracefulStop)
	go receiveTasks(stream, waitC, gracefulStop)

	<-waitC
	return nil
}

func hearbeat(stream v1.BuilderService_BreakRoomClient, gracefulStop chan os.Signal) {
	for {
		select {
		case <-gracefulStop:
			log.Println("stopping heartbeat")
			stream.CloseSend()
			return
		default:
			err := stream.Send(&v1.Heartbeat{
				Timestamp: ptypes.TimestampNow(),
			})
			if err != nil {
				log.Println(err)
			}
			time.Sleep(2 * time.Second)
		}
	}

}

func receiveTasks(stream v1.BuilderService_BreakRoomClient, waitC chan struct{}, gracefulStop chan os.Signal) {
	for {
		select {
		case <-gracefulStop:
			log.Println("stopping recieveTasks")
			return
		default:
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
}
