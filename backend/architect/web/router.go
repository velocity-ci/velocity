package web

import (
	"os"

	"github.com/dgrijalva/jwt-go"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func AddRoutes(e *echo.Echo) {
	e.POST("/v1/auth", createAuth)

	// Authenticated routes
	jwtConfig := middleware.JWTConfig{
		Claims:     &jwt.StandardClaims{},
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
	}
	r := e.Group("/v1/ssh")
	r.Use(middleware.JWTWithConfig(jwtConfig))
	r.POST("/known-hosts", createKnownHost)
	r.GET("/known-hosts", listKnownHost)

	r = e.Group("/v1/projects")
	r.Use(middleware.JWTWithConfig(jwtConfig))
	r.POST("", createProject)
}
