package rest

import (
	"net/http"

	"github.com/labstack/echo"
)

func health(c echo.Context) error {
	c.JSON(http.StatusOK, "OK")
	return nil
}
