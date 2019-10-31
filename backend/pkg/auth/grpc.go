package auth

import (
	"context"
	"fmt"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var (
	jwtSub = contextKey("jwt.sub")
)

// GRPCInterceptor intercepts GRPC requests and puts JWT subject into the request context.
func GRPCInterceptor(ctx context.Context) (context.Context, error) {
	if m, _ := grpc.Method(ctx); allowedMethod(m) {
		return ctx, nil
	}
	bearer, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(bearer, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	return context.WithValue(ctx, jwtSub, token.Claims.(jwt.MapClaims)["sub"]), nil
}

// GetSubject returns the subject from the given context
func GetSubject(ctx context.Context) (string, error) {
	if sub, ok := ctx.Value(jwtSub).(string); ok {
		return sub, nil
	}
	return "", fmt.Errorf("bad subject")
}

func allowedMethod(m string) bool {
	for _, s := range []string{
		"/velocity.v1.BuilderService/Register",
		"/velocity.v1.ProjectService/CreateProject",
	} {
		if m == s {
			return true
		}
	}
	return false
}
