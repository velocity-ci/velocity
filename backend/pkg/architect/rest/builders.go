package rest

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/auth"
	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
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
	broker         *broker
}

func newBuilderHandler(builderManager *builder.Manager, broker *broker) *builderHandler {
	return &builderHandler{
		builderManager: builderManager,
		broker:         broker,
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

type registerBuilderRequest struct {
	Secret string `json:"secret"`
}
type registerBuilderResponse struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

func newRegisterBuilderResponse(b *builder.Builder) *registerBuilderResponse {
	sessionDuration := time.Hour * 24 * 2
	token, _ := auth.NewJWT(sessionDuration, auth.AudienceBuilder, b.ID)

	return &registerBuilderResponse{
		ID:    b.ID,
		Token: token,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *builderHandler) register(c echo.Context) error {
	rB := new(registerBuilderRequest)
	if err := c.Bind(rB); err != nil {
		c.JSON(http.StatusBadRequest, "invalid payload")
		c.Logger().Warn(err)
		return nil
	}

	if rB.Secret != os.Getenv("BUILDER_SECRET") {
		c.JSON(http.StatusUnauthorized, "")
		return nil
	}

	b := h.builderManager.CreateBuilder()

	c.JSON(http.StatusCreated, newRegisterBuilderResponse(b))
	return nil
}

func (h *builderHandler) connect(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		velocity.GetLogger().Error("could not upgrade builder http connection", zap.Error(err))
		return nil
	}

	client := NewClient(ws)
	h.broker.save(client)

	go h.broker.monitor(client)

	return nil
}
