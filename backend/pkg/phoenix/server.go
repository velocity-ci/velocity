package phoenix

import (
	"fmt"
	"os"
	"sync"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/auth"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

type Server struct {
	ID string

	Socket *Socket

	// string:bool (set)
	subscribedTopics sync.Map

	authFunc func(*Server, *jwt.Token, string) error
}

func NewServer(
	ws *websocket.Conn,
	auth func(*Server, *jwt.Token, string) error,
	customEvents map[string]func(*PhoenixMessage) error,
	interlock bool) *Server {
	s := &Server{
		ID:       uuid.NewV4().String(),
		authFunc: auth,
	}

	customEvents[PhxJoinEvent] = s.subscribe
	customEvents[PhxLeaveEvent] = s.unsubscribe

	s.Socket = NewSocket(ws, customEvents, interlock)
	return s
}

func jwtKeyFunc(t *jwt.Token) (interface{}, error) {
	// Check the signing method (from echo.labstack.jwt middleware)
	if t.Method.Alg() != auth.JWTSigningMethod.Name {
		return nil, fmt.Errorf("Unexpected jwt signing method=%v", t.Header["alg"])
	}
	return []byte(os.Getenv("JWT_SECRET")), nil
}

func (s *Server) subscribe(m *PhoenixMessage) error {
	velocity.GetLogger().Debug("got subscribe", zap.String("topic", m.Topic))
	topic := m.Topic
	ref := m.Ref
	payload := m.Payload.(*PhoenixGuardianJoinPayload)
	token, err := jwt.ParseWithClaims(payload.Token, auth.JWTStandardClaims, jwtKeyFunc)
	if err != nil {
		s.Socket.Send(&PhoenixMessage{
			Event: PhxReplyEvent,
			Topic: topic,
			Ref:   ref,
			Payload: PhoenixReplyPayload{
				Status: "error",
				Response: map[string]string{
					"message": "access denied",
				},
			},
		}, false)
		velocity.GetLogger().Warn("could not authenticate client to channel", zap.String("serverID", s.ID), zap.Error(err))
		return err
	}
	if err := s.authFunc(s, token, topic); err != nil {
		s.Socket.Send(&PhoenixMessage{
			Event: PhxReplyEvent,
			Topic: topic,
			Ref:   ref,
			Payload: PhoenixReplyPayload{
				Status: "error",
				Response: map[string]string{
					"message": "access denied",
				},
			},
		}, false)
		velocity.GetLogger().Warn("could not authenticate client to channel", zap.String("serverID", s.ID), zap.Error(err))
		return err
	}
	s.subscribedTopics.Store(topic, true)
	s.Socket.ReplyOK(m)
	velocity.GetLogger().Debug("subscribed to", zap.String("topic", m.Topic))
	return nil
}

func (s *Server) unsubscribe(m *PhoenixMessage) error {
	topic := m.Topic
	if _, ok := s.subscribedTopics.Load(topic); ok {
		s.subscribedTopics.Delete(topic)
	}
	s.Socket.ReplyOK(m)

	return nil
}
