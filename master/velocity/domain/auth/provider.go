package auth

import (
	"os"
	"time"

	"github.com/VJftw/velocity/master/velocity/domain"
	jwt "github.com/dgrijalva/jwt-go"
)

type velocityClaims struct {
	Userame string `json:"username`
	jwt.StandardClaims
}

func NewAuthToken(user *domain.User) *domain.UserAuth {
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
