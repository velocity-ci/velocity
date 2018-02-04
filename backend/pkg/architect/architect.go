package architect

import (
	"context"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/velocity-ci/velocity/backend/pkg/architect/rest"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
)

type architect struct {
	server *echo.Echo
}

func (a *architect) Start() {
	a.server.Start(":8080")
}

func (a *architect) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return a.server.Shutdown(ctx)
}

type App interface {
	Start()
	Stop() error
}

func New() App {
	a := &architect{
		server: echo.New(),
	}

	a.server.Use(middleware.Logger())
	a.server.Use(middleware.Recover())

	validator, trans := domain.NewValidator()
	db := domain.NewStormDB("architect.db")

	rest.AddRoutes(
		a.server,
		db,
		validator,
		trans,
	)

	return a
}
