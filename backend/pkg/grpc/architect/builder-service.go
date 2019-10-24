package architect

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/auth"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
)

// BuilderServer implements the GRPC Builder Service
type BuilderServer struct {
}

// NewBuilderServer returns a new GRPC Builder service
func NewBuilderServer() *BuilderServer {
	return &BuilderServer{}
}

// Register registers a new builder, returning a token to authenticate with
func (s *BuilderServer) Register(context.Context, *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	id := uuid.NewV4()
	jwt, _ := auth.NewJWT(time.Hour, id.String())
	return &v1.RegisterResponse{
		Token: jwt,
	}, nil
}

// BreakRoom collects builders and schedules tasks on them
func (s *BuilderServer) BreakRoom(stream v1.BuilderService_BreakRoomServer) error {
	ctx := stream.Context()
	sub, err := auth.GetSubject(ctx)
	if err != nil {
		return err
	}
	sendChan := make(chan *v1.BreakResponse)
	heartbeatChan := make(chan *v1.Heartbeat)
	go breakRoomSend(ctx, stream, sendChan)
	go breakRoomHeartbeat(ctx, sub, heartbeatChan, sendChan)
	for {
		select {
		case <-ctx.Done():
			log.Println("stopping BreakRoom")
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

		heartbeatChan <- in

		time.Sleep(100 * time.Millisecond)
	}
}

// PushLogs persists logs from builders and updates stream state which in turn updates step, task and build states
func (s *BuilderServer) PushLogs(context.Context, *v1.LogsRequest) (*v1.LogsResponse, error) {
	return nil, nil
}

func breakRoomHeartbeat(
	ctx context.Context,
	sub string,
	heartbeatChan chan *v1.Heartbeat,
	sendChan chan *v1.BreakResponse,
) {
	for {
		select {
		case <-ctx.Done():
			log.Println("stopping breakRoomHeartbeat")
			close(heartbeatChan)
			return
		case hbt := <-heartbeatChan:
			t, _ := ptypes.Timestamp(hbt.GetTimestamp())
			log.Printf("recieved heartbeat from %s; %s", sub, t)
			sendChan <- &v1.BreakResponse{
				Type:      v1.BreakResponse_HEARTBEAT,
				Timestamp: ptypes.TimestampNow(),
				Data: &v1.BreakResponse_Heartbeat{
					Heartbeat: true,
				},
			}
		}

	}
}

func breakRoomSend(
	ctx context.Context,
	stream v1.BuilderService_BreakRoomServer,
	sendChan chan *v1.BreakResponse,
) {
	for {
		select {
		case <-ctx.Done():
			log.Println("stopping breakRoomSend")
			close(sendChan)
			return
		default:
			breakResponse := <-sendChan
			stream.Send(breakResponse)
		}
	}
}
