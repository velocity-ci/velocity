package web

import (
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/architect/domain"
	"github.com/velocity-ci/velocity/backend/architect/domain/persistence"
)

type RequestAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResponseAuth struct {
	Username string    `json:"username"`
	Token    string    `json:"token"`
	Expires  time.Time `json:"expires"`
}

func NewResponseAuth(u *domain.User) *ResponseAuth {
	now := time.Now()
	expires := time.Now().Add(time.Hour * 24 * 2)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: expires.Unix(),
		Issuer:    "Velocity CI",
		NotBefore: now.Unix(),
	})
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return &ResponseAuth{
		Username: u.Username,
		Token:    tokenString,
		Expires:  expires,
	}
}

func createAuth(c echo.Context) (err error) {
	rU := new(RequestAuth)
	if err = c.Bind(rU); err != nil {
		c.JSON(http.StatusBadRequest, "invalid payload")
		c.Logger().Warn(err)
		return nil
	}
	if _, err := domain.NewUser(rU.Username, rU.Password); err != nil {
		c.JSON(http.StatusBadRequest, domain.NewErrorMap(err))
		return nil
	}
	u, err := persistence.GetUser(domain.User{
		Username: rU.Username,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, nil)
		return nil
	}
	if !u.ValidatePassword(rU.Password) {
		c.JSON(http.StatusUnauthorized, nil)
		return nil
	}
	r := NewResponseAuth(u)
	c.JSON(http.StatusCreated, r)
	return nil
}
