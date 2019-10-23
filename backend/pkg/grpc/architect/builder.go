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

type BuilderServer struct{}

func NewBuilderServer() *BuilderServer {
	return &BuilderServer{}
}

func (s *BuilderServer) Register(context.Context, *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	id := uuid.NewV4()
	jwt, _ := auth.NewJWT(time.Hour, id.String())
	return &v1.RegisterResponse{
		Token: jwt,
	}, nil
}

func (s *BuilderServer) BreakRoom(stream v1.BuilderService_BreakRoomServer) error {
	sub, err := auth.GetSubject(stream.Context())
	if err != nil {
		return err
	}
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		t, _ := ptypes.Timestamp(in.GetTimestamp())

		log.Printf("recieved heartbeat from %s; %s", sub, t)
		stream.Send(&v1.BreakResponse{
			Timestamp: ptypes.TimestampNow(),
		})

		time.Sleep(100 * time.Millisecond)
	}
}

func (s *BuilderServer) PushLogs(context.Context, *v1.LogsRequest) (*v1.LogsResponse, error) {
	return nil, nil
}
