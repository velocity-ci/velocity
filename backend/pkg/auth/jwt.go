package auth

import (
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var JWTSigningMethod = jwt.SigningMethodHS512
var JWTStandardClaims = &jwt.StandardClaims{
	Issuer: "Velocity CI",
}

// Audience constants
const (
	AudienceUser    = "user"
	AudienceBuilder = "builder"
)

func NewJWT(expiryDuration time.Duration, audience, subject string) (string, time.Time) {
	now := time.Now()
	expires := time.Now().Add(expiryDuration)

	claims := JWTStandardClaims
	claims.ExpiresAt = expires.Unix()
	claims.NotBefore = now.Unix()
	claims.IssuedAt = now.Unix()
	claims.Subject = subject
	claims.Audience = audience

	token := jwt.NewWithClaims(JWTSigningMethod, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return tokenString, expires
}
