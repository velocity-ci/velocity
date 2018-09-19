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
		sentUnacked:      map[uint64]*wsMessage{},
		recievedUnacked:  map[uint64]*PhoenixMessage{},
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
	sentUnacked     map[uint64]*wsMessage
	recievedUnacked map[uint64]*PhoenixMessage
	mLock           sync.Mutex

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
			velocity.GetLogger().Debug("sent message", zap.String("event", m.Event), zap.String("topic", m.Topic), zap.Uint64("ref", m.Ref))
			s.mLock.Lock()
			if m.Event == PhxReplyEvent || m.Event == PhxErrorEvent {
				delete(s.recievedUnacked, m.Ref)
			} else {
				s.sentUnacked[m.Ref].sent = time.Now()
			}
			s.messageQueue = s.messageQueue[1:]
			s.mLock.Unlock()
		}
		time.Sleep(10 * time.Millisecond)
	}
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
		s.mLock.Lock()
		s.sentUnacked[ref] = &wsMessage{
			sent:    time.Now(),
			message: &m,
		}
		s.mLock.Unlock()
		s.lastHeartbeatRef = ref
		s.wsWriteLock.Lock()
		s.ws.WriteJSON(&m)
		s.wsWriteLock.Unlock()

		time.Sleep(5 * time.Second)
	}
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
	s.mLock.Lock()
	if s.interlock && m.Event != PhxReplyEvent {
		s.sentUnacked[m.Ref] = &qM
	}
	s.messageQueue = append(s.messageQueue, m)
	s.mLock.Unlock()

	if sync {
		return <-qM.response
	}

	return nil
}

func (s *Socket) monitor() {
	for s.connected {
		m := &PhoenixMessage{}
		err := s.ws.ReadJSON(m)
		velocity.GetLogger().Debug("RECIEVED MESSAGE", zap.String("message", m.Event))
		if err != nil {
			velocity.GetLogger().Error("could not read websocket message", zap.Error(err))
			s.ws.Close()
			s.connected = false
			break
		}

		if eventFunc, ok := s.customEvents[m.Event]; ok {
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
			}
		}
	}
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
	if _, ok := s.sentUnacked[m.Ref]; ok {
		if m.Ref == s.lastHeartbeatRef {
			s.lastHeartbeatRef = 0
			velocity.GetLogger().Debug("heartbeat pong", zap.Uint64("ref", m.Ref), zap.Duration("latency", time.Now().Sub(s.sentUnacked[m.Ref].sent)))
			s.healthy = true
			// requeue
			for ref, m := range s.sentUnacked {
				if !m.sent.IsZero() && time.Now().Sub(m.sent) > 5*time.Second {
					s.sentUnacked[ref].sent = time.Time{}
					s.messageQueue = append(s.messageQueue, m.message)
					velocity.GetLogger().Debug("requeued", zap.Uint64("ref", ref))
				}
			}
		}
		if s.sentUnacked[m.Ref].response != nil {
			s.sentUnacked[m.Ref].response <- m.Payload.(*PhoenixReplyPayload)
			close(s.sentUnacked[m.Ref].response)
		}
		delete(s.sentUnacked, m.Ref)
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
