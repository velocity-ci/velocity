package web

import (
	"net/http"

	"github.com/velocity-ci/velocity/backend/architect/domain"
	"github.com/velocity-ci/velocity/backend/architect/domain/persistence"

	"github.com/labstack/echo"
)

type RequestKnownHost struct {
	Entry string `json:"entry"`
}

type ResponseKnownHost struct {
	ID                string   `json:"id"`
	Hosts             []string `json:"hosts"`
	Comment           string   `json:"comment"`
	SHA256Fingerprint string   `json:"sha256"`
	MD5Fingerprint    string   `json:"md5"`
}

type ListKnownHost struct {
	Total int                  `json:"total"`
	Data  []*ResponseKnownHost `json:"data"`
}

func NewResponseKnownHost(k *domain.KnownHost) *ResponseKnownHost {
	return &ResponseKnownHost{
		ID:                k.UUID,
		Hosts:             k.Hosts,
		Comment:           k.Comment,
		SHA256Fingerprint: k.SHA256Fingerprint,
		MD5Fingerprint:    k.MD5Fingerprint,
	}
}

func createKnownHost(c echo.Context) error {
	rKH := new(RequestKnownHost)
	if err := c.Bind(rKH); err != nil {
		c.JSON(http.StatusBadRequest, "invalid payload")
		c.Logger().Warn(err)
		return nil
	}
	k, err := domain.NewKnownHost(rKH.Entry)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorMap(err))
		return nil
	}

	if err := persistence.SaveKnownHost(k); err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return nil
	}

	c.JSON(http.StatusCreated, NewResponseKnownHost(k))
	return nil
}

func listKnownHost(c echo.Context) error {
	ks, total := persistence.GetKnownHosts()

	respData := []*ResponseKnownHost{}
	for _, k := range ks {
		respData = append(respData, NewResponseKnownHost(k))
	}

	c.JSON(http.StatusOK, ListKnownHost{
		Total: total,
		Data:  respData,
	})

	return nil
}
