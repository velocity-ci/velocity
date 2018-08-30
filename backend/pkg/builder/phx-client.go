package builder

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

type PhoenixWSClient struct {
	ws   *websocket.Conn
	lock sync.RWMutex

	address string
	headers http.Header
}

func NewPhoenixWSClient(address string) *PhoenixWSClient {
	ws := &PhoenixWSClient{
		address: address,
	}

	return ws
}

func (sws *PhoenixWSClient) Subscribe(topic, token string) error {

}

// func (ws *PhoenixWSClient) Start() error {
// }

func (sws *PhoenixWSClient) WriteJSON(m interface{}) error {
	sws.lock.Lock()
	defer sws.lock.Unlock()

	return sws.ws.WriteJSON(m)
}

func (ws *PhoenixWSClient) dial() error {
	var dialer *websocket.Dialer
	conn, _, err := dialer.Dial(
		ws.address,
		ws.headers,
	)
	ws.ws = conn
	return err
}

func (ws *PhoenixWSClient) monitor() error {

	for {

		command := &builder.BuilderCtrlMessage{}
		err := ws.ws.ReadJSON(command)
		if err != nil {
			velocity.GetLogger().Error("could not read websocket message", zap.Error(err))
			ws.ws.Close()
		}

		if command.Command == builder.CommandBuild {
			velocity.GetLogger().Info("got build", zap.Any("payload", command.Payload))
			runBuild(command.Payload.(*builder.BuildCtrl), ws)
		} else if command.Command == builder.CommandKnownHosts {
			velocity.GetLogger().Info("got known hosts", zap.Any("payload", command.Payload))
			updateKnownHosts(command.Payload.(*builder.KnownHostCtrl))
		}
	}

}
