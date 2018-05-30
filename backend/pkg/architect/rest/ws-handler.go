package rest

import (
	"github.com/golang/glog"
	"github.com/labstack/echo"
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
	// auth := c.Request().Header.Get("Authorization")
	// if auth == "" {
	// 	c.JSON(http.StatusUnauthorized, "")
	// 	return nil
	// }
	// if auth != os.Getenv("BUILDER_TOKEN") {
	// 	c.JSON(http.StatusUnauthorized, "")
	// 	return nil
	// }

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		glog.Error(err)
		return nil
	}

	client := NewClient(ws)
	h.broker.save(client)

	go h.broker.monitor(client)
	return nil
}
