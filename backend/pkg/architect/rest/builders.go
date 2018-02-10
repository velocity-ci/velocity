package rest

import (
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
)

type builderResponse struct {
	ID        string    `json:"id"`
	State     string    `json:"state"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type builderList struct {
	Total int                `json:"total"`
	Data  []*builderResponse `json:"data"`
}

func newBuilderResponse(b *builder.Builder) *builderResponse {
	return &builderResponse{
		ID:        b.ID,
		State:     b.State,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

type builderHandler struct {
	builderManager *builder.Manager
}

func newBuilderHandler(builderManager *builder.Manager) *builderHandler {
	return &builderHandler{
		builderManager: builderManager,
	}
}

func (h *builderHandler) getAll(c echo.Context) error {
	pQ := getPagingQueryParams(c)
	if pQ == nil {
		return nil
	}

	bs, total := h.builderManager.GetAll(pQ)
	rBuilders := []*builderResponse{}
	for _, b := range bs {
		rBuilders = append(rBuilders, newBuilderResponse(b))
	}

	c.JSON(http.StatusOK, builderList{
		Total: total,
		Data:  rBuilders,
	})
	return nil
}

func (h *builderHandler) getByID(c echo.Context) error {

	if b := getBuilderByID(c, h.builderManager); b != nil {
		c.JSON(http.StatusOK, newBuilderResponse(b))
	}

	return nil
}

func getBuilderByID(c echo.Context, bM *builder.Manager) *builder.Builder {
	id := c.Param("id")

	b, err := bM.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, "not found")
		return nil
	}

	return b
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (h *builderHandler) connect(c echo.Context) error {
	auth := c.Request().Header.Get("Authorization")
	logrus.Debugf("builder authorization attempt with: %s", auth)
	if auth == "" {
		c.JSON(http.StatusUnauthorized, "")
		return nil
	}
	if auth != os.Getenv("BUILDER_SECRET") {
		c.JSON(http.StatusUnauthorized, "")
		return nil
	}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logrus.Error(err)
		return nil
	}

	h.builderManager.CreateBuilder(ws)
	return nil
}
