package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/unrolled/render"
)

// JWT -
type JWT struct {
	render *render.Render
}

// NewJWT -
func NewJWT(renderer *render.Render) *JWT {
	return &JWT{
		render: renderer,
	}
}

type key string

const requestUsername key = "username"

// ServeHTTP -
func (m *JWT) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	tokenString, err := fromAuthHeader(r)
	if err != nil {
		tokenString, err = fromQueryString(r)
	}

	if err != nil {
		m.render.JSON(rw, http.StatusUnauthorized, nil)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, m.render.JSON(rw, http.StatusUnauthorized, nil)
			// return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	ctx := r.Context()
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// fmt.Println(claims)
		ctx = context.WithValue(ctx, requestUsername, claims["username"])
	} else {
		fmt.Println(err)
		return
	}
	next(rw, r.WithContext(ctx))
}

func fromQueryString(r *http.Request) (string, error) {
	authToken := r.URL.Query().Get("authToken")
	if authToken == "" {
		return "", errors.New("Missing authToken")
	}

	return authToken, nil
}

// FromAuthHeader is a "TokenExtractor" that takes a give request and extracts
// the JWT token from the Authorization header.
func fromAuthHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Missing Token")
	}

	// TODO: Make this a bit more robust, parsing-wise
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return "", fmt.Errorf("Authorization header format must be Bearer {token}")
	}

	return authHeaderParts[1], nil
}

// UsernameFromContext - returns the auth token association with a context.
func UsernameFromContext(ctx context.Context) string {
	return ctx.Value(requestUsername).(string)
}
