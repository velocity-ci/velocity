package phoenix

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

type wsMessage struct {
	sent     time.Time
	message  *PhoenixMessage
	response chan *PhoenixReplyPayload
}

type PhoenixWSClient struct {
	lock             sync.RWMutex
	ws               *websocket.Conn
	connected        bool
	lastHeartbeatRef uint64

	unacked      map[uint64]*wsMessage
	messageQueue []uint64
	refCounter   uint64

	subscribedTopics map[string]PhoenixGuardianJoinPayload

	customEvents map[string]func(*PhoenixMessage) error

	address string
	headers http.Header
}

func NewPhoenixWSClient(address string, customEvents map[string]func(*PhoenixMessage) error) (*PhoenixWSClient, error) {
	ws := &PhoenixWSClient{
		address:          address,
		unacked:          map[uint64]*wsMessage{},
		messageQueue:     []uint64{},
		lastHeartbeatRef: 0,
		connected:        false,
		refCounter:       0,
		subscribedTopics: map[string]PhoenixGuardianJoinPayload{},
		customEvents:     customEvents,
	}

	err := ws.dial()
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (sws *PhoenixWSClient) Wait(reconnects int) {
	for i := 0; i <= reconnects-1; i++ {
		for sws.connected {
			time.Sleep(1 * time.Second)
		}
		sws.dial()
	}
}

func (sws *PhoenixWSClient) dial() error {
	var dialer *websocket.Dialer
	conn, _, err := dialer.Dial(
		sws.address,
		sws.headers,
	)
	if err != nil {
		return err
	}
	sws.connected = true
	sws.ws = conn
	go sws.monitor()
	go sws.heartbeat()
	go sws.worker()
	for topic, payload := range sws.subscribedTopics {
		err := sws.Subscribe(topic, payload.Token)
		if err != nil {
			velocity.GetLogger().Warn("could not resubscribe", zap.String("topic", topic))
		}
	}
	return nil
}

func (sws *PhoenixWSClient) worker() {
	for sws.connected {
		if len(sws.messageQueue) > 0 {
			sws.lock.Lock()
			sws.ws.WriteJSON(sws.unacked[sws.messageQueue[0]].message)
			sws.unacked[sws.messageQueue[0]].sent = time.Now()
			sws.messageQueue = sws.messageQueue[1:]
			sws.lock.Unlock()
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (sws *PhoenixWSClient) monitor() {
	for sws.connected {
		m := &PhoenixMessage{}
		err := sws.ws.ReadJSON(m)
		if err != nil {
			velocity.GetLogger().Error("could not read websocket message", zap.Error(err))
			sws.ws.Close()
			sws.connected = false
			break
		}

		if m.Event == PhxReplyEvent {
			if _, ok := sws.unacked[m.Ref]; ok {
				if m.Ref == sws.lastHeartbeatRef {
					sws.lastHeartbeatRef = 0
					velocity.GetLogger().Debug("heartbeat pong", zap.Uint64("ref", m.Ref), zap.Duration("latency", time.Now().Sub(sws.unacked[m.Ref].sent)))
					// requeue
					for ref, m := range sws.unacked {
						if !m.sent.IsZero() && time.Now().Sub(m.sent) > 5*time.Second {
							sws.unacked[ref].sent = time.Time{}
							sws.messageQueue = append(sws.messageQueue, ref)
							velocity.GetLogger().Debug("requeued", zap.Uint64("ref", ref))
						}
					}
				}
				if sws.unacked[m.Ref].response != nil {
					sws.unacked[m.Ref].response <- m.Payload.(*PhoenixReplyPayload)
					close(sws.unacked[m.Ref].response)
				}
				delete(sws.unacked, m.Ref)
			} else {
				velocity.GetLogger().Warn("message not unacked", zap.Uint64("ref", m.Ref), zap.String("event", m.Event), zap.String("topic", m.Topic))
			}
		} else if eventFunc, ok := sws.customEvents[m.Event]; ok {
			if err := eventFunc(m); err != nil {
				velocity.GetLogger().Error("error in custom event", zap.Error(err))
			}
		}
	}
}

func (sws *PhoenixWSClient) getNewRef() uint64 {
	sws.refCounter++
	return sws.refCounter
}

func (sws *PhoenixWSClient) heartbeat() {
	for sws.connected {
		if sws.lastHeartbeatRef != 0 {
			velocity.GetLogger().Warn("heartbeat timed out", zap.Uint64("ref", sws.lastHeartbeatRef))
			sws.connected = false
			break
		}
		sws.lock.Lock()
		ref := sws.getNewRef()
		m := PhoenixMessage{
			Event:   PhxHeartbeatEvent,
			Topic:   PhxSystemTopic,
			Ref:     ref,
			Payload: map[string]string{},
		}
		sws.unacked[ref] = &wsMessage{
			sent:    time.Now(),
			message: &m,
		}
		sws.lastHeartbeatRef = ref
		sws.ws.WriteJSON(&m)
		// velocity.GetLogger().Debug("heartbeat ping", zap.Uint64("ref", ref))
		sws.lock.Unlock()
		time.Sleep(5 * time.Second)
	}
}

func (sws *PhoenixWSClient) Send(m *PhoenixMessage, sync bool) *PhoenixReplyPayload {
	// 1. create new wsmessage with ref no, create response chan if sync
	// 2. add to unacked and message queue
	// if sync,
	// 3. wait for response chan, return response payload
	m.Ref = sws.getNewRef()
	qM := wsMessage{message: m}
	if sync {
		qM.response = make(chan *PhoenixReplyPayload)
	}
	sws.unacked[m.Ref] = &qM
	sws.messageQueue = append(sws.messageQueue, m.Ref)

	if sync {
		return <-qM.response
	}

	return nil
}

func (sws *PhoenixWSClient) Subscribe(topic, token string) error {
	velocity.GetLogger().Debug("subscribing to", zap.String("topic", topic), zap.Int("token(len)", len(token)))
	resp := sws.Send(&PhoenixMessage{
		Event: PhxJoinEvent,
		Topic: topic,
		Payload: PhoenixGuardianJoinPayload{
			Token: token,
		},
	}, true)

	if resp.Status != ResponseOK {
		velocity.GetLogger().Error("could not subscribe", zap.String("topic", topic), zap.Int("token(len)", len(token)))
		return fmt.Errorf("%v", resp.Response)
	}

	velocity.GetLogger().Debug("subscribed to", zap.String("topic", topic))

	sws.subscribedTopics[topic] = PhoenixGuardianJoinPayload{
		Token: token,
	}

	return nil
}
