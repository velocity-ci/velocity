package rest

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

type websocketHandler struct {
	broker *broker
}

func newWebsocketHandler(broker *broker) *websocketHandler {
	return &websocketHandler{
		broker: broker,
	}
}

func (h *websocketHandler) phxClient(c echo.Context) error {

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		velocity.GetLogger().Error("could not upgrade client websocket", zap.Error(err))
		return nil
	}

	client := phoenix.NewServer(
		ws,
		func(*phoenix.Server, *jwt.Token, string) error {
			return nil
		},
		map[string]func(*phoenix.PhoenixMessage) error{},
		false,
	)
	h.broker.save(client)

	return nil
}
