package rest

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"
)

type knownHostRequest struct {
	Entry string `json:"entry"`
}

type knownHostResponse struct {
	ID                string   `json:"id"`
	Hosts             []string `json:"hosts"`
	Comment           string   `json:"comment"`
	SHA256Fingerprint string   `json:"sha256"`
	MD5Fingerprint    string   `json:"md5"`
}

type knownhostList struct {
	Total int                  `json:"total"`
	Data  []*knownHostResponse `json:"data"`
}

func newKnownHostResponse(k *knownhost.KnownHost) *knownHostResponse {
	return &knownHostResponse{
		ID:                k.ID,
		Hosts:             k.Hosts,
		Comment:           k.Comment,
		SHA256Fingerprint: k.SHA256Fingerprint,
		MD5Fingerprint:    k.MD5Fingerprint,
	}
}

type knownHostHandler struct {
	knownHostManager *knownhost.Manager
}

func newKnownHostHandler(knownHostManager *knownhost.Manager) *knownHostHandler {
	return &knownHostHandler{
		knownHostManager: knownHostManager,
	}
}

func (h *knownHostHandler) create(c echo.Context) error {
	rKH := new(knownHostRequest)
	if err := c.Bind(rKH); err != nil {
		c.JSON(http.StatusBadRequest, "invalid payload")
		c.Logger().Warn(err)
		return nil
	}
	k, err := h.knownHostManager.Create(rKH.Entry)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.ErrorMap)
		return nil
	}

	c.JSON(http.StatusCreated, newKnownHostResponse(k))
	return nil
}
