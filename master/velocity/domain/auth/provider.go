package auth

import (
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/velocity-ci/velocity/master/velocity/domain"
)

type velocityClaims struct {
	Userame string `json:"username`
	jwt.StandardClaims
}

func NewAuthToken(user *domain.BoltUser) *domain.UserAuth {
	now := time.Now()
	expires := time.Now().Add(time.Hour * 24 * 2)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, velocityClaims{
		user.Username,
		jwt.StandardClaims{
			ExpiresAt: expires.Unix(),
			Issuer:    "Velocity",
			NotBefore: now.Unix(),
		},
	})
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return &domain.UserAuth{
		Username: user.Username,
		Token:    tokenString,
		Expires:  expires,
	}
}
