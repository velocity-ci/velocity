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
func BreakRoom(address string, token string, stop chan os.Signal) error {
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

	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		<-stop
		cancelFunc()
	}()

	stream, err := client.BreakRoom(ctx)
	if err != nil {
		return err
	}

	sendChan := make(chan *v1.Heartbeat)
	go breakRoomSend(ctx, stream, sendChan)
	go breakRoomHeartbeat(ctx, sendChan)

	for {
		select {
		case <-ctx.Done():
			log.Println("stopping recieveTasks")
			return ctx.Err()
		default:
		}
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		switch in.GetType() {
		case v1.BreakResponse_HEARTBEAT:
			handleHeartbeat(ctx, in)
		case v1.BreakResponse_TASK:
			handleTask(ctx, in)
		}
	}
}

func breakRoomSend(
	ctx context.Context,
	stream v1.BuilderService_BreakRoomClient,
	sendChan chan *v1.Heartbeat,
) {
	for {
		select {
		case <-ctx.Done():
			log.Println("stopping breakRoomSend")
			close(sendChan)
			stream.CloseSend()
			return
		default:
			breakResponse := <-sendChan
			stream.Send(breakResponse)
		}
	}
}

func breakRoomHeartbeat(ctx context.Context, sendChan chan *v1.Heartbeat) {
	for {
		select {
		case <-ctx.Done():
			log.Println("stopping heartbeat")
			return
		default:
			sendChan <- &v1.Heartbeat{
				Timestamp: ptypes.TimestampNow(),
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func handleHeartbeat(ctx context.Context, resp *v1.BreakResponse) error {
	t, _ := ptypes.Timestamp(resp.Timestamp)
	log.Printf("Got heartbeat at %+v", t)
	return nil
}

func handleTask(ctx context.Context, resp *v1.BreakResponse) error {
	task := resp.GetTask()
	log.Printf("Got task %s: %+v", task.GetId(), resp)

	return nil
}
