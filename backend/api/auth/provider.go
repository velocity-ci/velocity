package auth

import (
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type VelocityClaims struct {
	Userame string `json:"username"`
	jwt.StandardClaims
}

// NewAuthToken - Returns a new AuthToken for a given name
func NewAuthToken(name string) *UserAuth {
	now := time.Now()
	expires := time.Now().Add(time.Hour * 24 * 2)
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, VelocityClaims{
		Userame: name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expires.Unix(),
			Issuer:    "Velocity",
			NotBefore: now.Unix(),
		},
	})
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return &UserAuth{
		Username: name,
		Token:    tokenString,
		Expires:  expires,
	}
}
