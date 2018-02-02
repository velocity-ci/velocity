package web

import (
	"context"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type App interface {
	Start()
	Stop()
}

type Web struct {
	Server *echo.Echo
}

func NewWeb() *Web {
	w := &Web{
		Server: echo.New(),
	}

	w.Server.Use(middleware.Logger())
	w.Server.Use(middleware.Recover())

	AddRoutes(w.Server)

	return w
}

func (w *Web) Start() {
	w.Server.Start(":8080")
}

func (w *Web) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := w.Server.Shutdown(ctx); err != nil {
		w.Server.Logger.Fatal(err)
	}
}
