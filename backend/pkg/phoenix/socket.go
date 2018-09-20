package phoenix

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

func NewSocket(ws *websocket.Conn, customEvents map[string]func(*PhoenixMessage) error, interlock bool) *Socket {
	s := &Socket{
		ws:               ws,
		connected:        true,
		messageQueue:     []*PhoenixMessage{},
		lastHeartbeatRef: 0,
		refCounter:       0,
		customEvents:     customEvents,
		interlock:        interlock,
	}

	velocity.GetLogger().Debug("new socket", zap.String("ws", ws.LocalAddr().String()), zap.Bool("interlocked", interlock))

	go s.monitor()
	if s.interlock {
		go s.heartbeat()
	} else {
		s.healthy = true
	}
	go s.worker()

	return s
}

type Socket struct {
	ws               *websocket.Conn
	wsWriteLock      sync.Mutex
	connected        bool
	healthy          bool
	lastHeartbeatRef uint64

	messageQueue    []*PhoenixMessage
	sentUnacked     sync.Map
	recievedUnacked sync.Map

	refCounter     uint64
	refCounterLock sync.Mutex

	customEvents map[string]func(*PhoenixMessage) error

	interlock bool
}

func (s Socket) IsConnected() bool {
	return s.connected
}

func (s *Socket) getNewRef() uint64 {
	s.refCounterLock.Lock()
	defer s.refCounterLock.Unlock()
	s.refCounter++
	return s.refCounter
}

func (s *Socket) worker() {
	for s.connected {
		if s.healthy && len(s.messageQueue) > 0 {
			s.wsWriteLock.Lock()
			m := s.messageQueue[0]
			s.ws.WriteJSON(m)
			s.wsWriteLock.Unlock()
			if m.Event == PhxReplyEvent || m.Event == PhxErrorEvent {
				s.recievedUnacked.Delete(m.Ref)
			} else {
				if x, ok := s.sentUnacked.Load(m.Ref); ok {
					if v, ok := x.(*wsMessage); ok {
						v.sent = time.Now()
						s.sentUnacked.Store(m.Ref, v)
					} else {
						velocity.GetLogger().Debug("invalid value found in map", zap.String("map", "sentUnacked"))
					}
				}
			}
			s.messageQueue = s.messageQueue[1:]
		}
		time.Sleep(10 * time.Millisecond)
	}
	velocity.GetLogger().Debug("worker ended", zap.String("remote", s.ws.RemoteAddr().String()))
}

func (s *Socket) heartbeat() {
	for s.connected {
		if s.lastHeartbeatRef != 0 {
			velocity.GetLogger().Warn("heartbeat timed out", zap.Uint64("ref", s.lastHeartbeatRef))
			s.connected = false
			break
		}
		ref := s.getNewRef()
		m := PhoenixMessage{
			Event:   PhxHeartbeatEvent,
			Topic:   PhxSystemTopic,
			Ref:     ref,
			Payload: map[string]string{},
		}
		s.sentUnacked.Store(ref, &wsMessage{
			sent:    time.Now(),
			message: &m,
		})
		s.lastHeartbeatRef = ref
		s.wsWriteLock.Lock()
		s.ws.WriteJSON(&m)
		s.wsWriteLock.Unlock()

		time.Sleep(5 * time.Second)
	}

	velocity.GetLogger().Debug("heartbeat ended", zap.String("remote", s.ws.RemoteAddr().String()))
}

func (s *Socket) Send(m *PhoenixMessage, sync bool) *PhoenixReplyPayload {
	// 1. create new wsmessage with ref no, create response chan if sync
	// 2. add to unacked and message queue
	// if sync,
	// 3. wait for response chan, return response payload
	// if _, ok := s.subscribedTopics[m.Topic]; !ok {
	// 	velocity.GetLogger().Error("not sending as not subscribed", zap.String("topic", m.Topic))
	// 	return nil
	// }
	if m.Ref < 1 {
		m.Ref = s.getNewRef()
	}
	qM := wsMessage{message: m}
	if sync {
		qM.response = make(chan *PhoenixReplyPayload)
	}
	if s.interlock && m.Event != PhxReplyEvent {
		s.sentUnacked.Store(m.Ref, &qM)
	}
	s.messageQueue = append(s.messageQueue, m)

	if sync {
		return <-qM.response
	}

	return nil
}

func (s *Socket) handleMessage(m *PhoenixMessage) {
	if eventFunc, ok := s.customEvents[m.Event]; ok {
		velocity.GetLogger().Debug("executing custom event", zap.String("event", m.Event))
		if err := eventFunc(m); err != nil {
			velocity.GetLogger().Error("error in custom event", zap.Error(err))
		}
	} else {
		switch m.Event {
		case PhxReplyEvent:
			s.handlePhxReplyEvent(m)
			break
		case PhxHeartbeatEvent:
			s.handlePhxHeartbeatEvent(m.Ref)
			break
		default:
			velocity.GetLogger().Warn("event not handled", zap.String("event", m.Event), zap.String("topic", m.Topic))
			break
		}
	}
}

func (s *Socket) monitor() {
	for s.connected {
		m := &PhoenixMessage{}
		err := s.ws.ReadJSON(m)
		if err != nil {
			velocity.GetLogger().Error("could not read websocket message", zap.Error(err))
			s.ws.Close()
			s.connected = false
			break
		}

		go s.handleMessage(m)
	}
	velocity.GetLogger().Debug("monitor ended", zap.String("remote", s.ws.RemoteAddr().String()))
}

func (s *Socket) ReplyOK(m *PhoenixMessage) {
	s.Send(&PhoenixMessage{
		Event: PhxReplyEvent,
		Topic: m.Topic,
		Ref:   m.Ref,
		Payload: PhoenixReplyPayload{
			Status:   "ok",
			Response: map[string]string{},
		},
	}, false)
}

func (s *Socket) handlePhxReplyEvent(m *PhoenixMessage) {
	if _, ok := s.sentUnacked.Load(m.Ref); ok {
		if m.Ref == s.lastHeartbeatRef {
			s.lastHeartbeatRef = 0
			if val, ok := s.sentUnacked.Load(m.Ref); ok {
				velocity.GetLogger().Debug("heartbeat pong", zap.Uint64("ref", m.Ref), zap.Duration("latency", time.Now().Sub(val.(*wsMessage).sent)))
			}
			s.healthy = true
			// requeue
			s.sentUnacked.Range(func(k, v interface{}) bool {
				if m, ok := v.(*wsMessage); ok {
					if !m.sent.IsZero() && time.Now().Sub(m.sent) > 5*time.Second {
						m.sent = time.Time{}
						s.sentUnacked.Store(k, m)
						s.messageQueue = append(s.messageQueue, m.message)

						velocity.GetLogger().Debug("requeued", zap.Uint64("ref", k.(uint64)))
					}
				}
				return true
			})
		}
		if val, ok := s.sentUnacked.Load(m.Ref); ok {
			if val, ok := val.(*wsMessage); ok {
				if val.response != nil {
					val.response <- m.Payload.(*PhoenixReplyPayload)
					close(val.response)
				}
				s.sentUnacked.Delete(m.Ref)
			}
		}
	} else {
		velocity.GetLogger().Warn("message not unacked (interlock is disabled?)", zap.Uint64("ref", m.Ref), zap.String("event", m.Event), zap.String("topic", m.Topic))
	}
}

func (s *Socket) handlePhxHeartbeatEvent(ref uint64) {
	s.wsWriteLock.Lock()
	s.ws.WriteJSON(PhoenixMessage{
		Event: PhxReplyEvent,
		Topic: PhxSystemTopic,
		Ref:   ref,
		Payload: PhoenixReplyPayload{
			Status:   ResponseOK,
			Response: map[string]string{},
		},
	})
	s.wsWriteLock.Unlock()
}
