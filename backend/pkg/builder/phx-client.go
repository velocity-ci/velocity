package builder

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/phoenix"

	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

type wsMessage struct {
	sent     time.Time
	message  *phoenix.PhoenixMessage
	response chan *phoenix.PhoenixReplyPayload
}

type PhoenixWSClient struct {
	ws   *websocket.Conn
	lock sync.RWMutex

	unacked      map[int]wsMessage
	messageQueue []int
	// wsMessage states: sent (unacked), resent (unacked), delivered (acked)
	// monitor func listens for:
	//  - reply ACKs to remove messages from unacked. CHECK if reply is OK - if not, print errors before discarding send resp over channel then close.
	//  - reply Heartbeats update last heartbeat and move all unacked for over x period into messageQueue.
	// Note: Backend is going to need to keep track of messages that it has received but not processed/acked yet.
	// heartbeat func
	// - periodically sends out new heartbeats and if last heartbeat is 3x older than period, assumes lost connection.
	// - along with new heartbeats,
	// - if lost connection, move all unacked back into messageQueue and reconnect
	// worker func
	// - takes messages from queue and sends them (sleep for 0.1 seconds?) TODO: implement debouncing of logs

	address string
	headers http.Header
}

func NewPhoenixWSClient(address string) *PhoenixWSClient {
	ws := &PhoenixWSClient{
		address: address,
	}

	return ws
}

func (sws *PhoenixWSClient) send(m *phoenix.PhoenixMessage, sync bool) *phoenix.PhoenixReplyPayload {
	// 1. create new wsmessage with ref no, create response chan if sync
	// 2. add to unacked and message queue
	// if sync,
	// 3. wait for response chan, return response payload
	m.Ref = 3
	qM := wsMessage{message: m}
	if sync {
		qM.response = make(chan *phoenix.PhoenixReplyPayload)
	}
	sws.unacked[3] = qM
	sws.messageQueue = append(sws.messageQueue, 3)

	if sync {
		return <-qM.response
	}

	return nil
}

func (sws *PhoenixWSClient) Subscribe(topic, token string) error {
	resp := sws.send(&phoenix.PhoenixMessage{
		Event: phoenix.PhxJoinEvent,
		Topic: topic,
		Payload: phoenix.PhoenixGuardianJoinPayload{
			Token: token,
		},
	}, true)

	if resp.Status != phoenix.ResponseOK {
		velocity.GetLogger().Error("could not subscribe", zap.String("topic", topic), zap.Int("token(len)", len(token)))
		return fmt.Errorf("%v", resp.Response)
	}

	return nil
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
