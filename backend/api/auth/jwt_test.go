package auth_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/unrolled/render"
	"github.com/velocity-ci/velocity/backend/api/auth"
)

func TestMiddlewareMissingToken(t *testing.T) {
	render := render.New()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, 200, "Woo")
	})

	jwt := auth.NewJWT(render)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwt.ServeHTTP(w, r, next)
	}))
	defer ts.Close()

	res, _ := http.Get(ts.URL)
	assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
}

func TestValidToken(t *testing.T) {
	os.Setenv("JWT_TOKEN", "test")
	render := render.New()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.Text(w, http.StatusOK, auth.UsernameFromContext(r.Context()))
	})

	jwt := auth.NewJWT(render)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwt.ServeHTTP(w, r, next)
	}))
	defer ts.Close()

	// From Authorization Header
	request, _ := http.NewRequest("GET", ts.URL, nil)
	authToken := auth.NewAuthToken("Bob")
	request.Header.Set("Authorization", fmt.Sprintf("bearer %s", authToken.Token))
	res, _ := ts.Client().Do(request)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	bodyBytes, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, "Bob", string(bodyBytes))

	// From QueryString
	request, _ = http.NewRequest("GET", fmt.Sprintf("%s?authToken=%s", ts.URL, authToken.Token), nil)
	res, _ = ts.Client().Do(request)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	bodyBytes, _ = ioutil.ReadAll(res.Body)
	assert.Equal(t, "Bob", string(bodyBytes))

}

func TestInvalidParseToken(t *testing.T) {
	os.Setenv("JWT_TOKEN", "test")
	render := render.New()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, http.StatusOK, "Woo")
	})

	jwt := auth.NewJWT(render)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwt.ServeHTTP(w, r, next)
	}))
	defer ts.Close()

	request, _ := http.NewRequest("GET", ts.URL, nil)
	authToken := auth.NewAuthToken("Bob")
	request.Header.Set("Authorization", fmt.Sprintf("bearer dd%s", authToken.Token))
	res, _ := ts.Client().Do(request)
	assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
}

func TestNoBearerToken(t *testing.T) {
	os.Setenv("JWT_TOKEN", "test")
	render := render.New()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, http.StatusOK, "Woo")
	})

	jwt := auth.NewJWT(render)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwt.ServeHTTP(w, r, next)
	}))
	defer ts.Close()

	request, _ := http.NewRequest("GET", ts.URL, nil)
	authToken := auth.NewAuthToken("Bob")
	request.Header.Set("Authorization", fmt.Sprintf("%s", authToken.Token))
	res, _ := ts.Client().Do(request)
	assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
}

func TestInvalidSigningMethodToken(t *testing.T) {
	os.Setenv("JWT_TOKEN", "test")
	render := render.New()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, http.StatusOK, "Woo")
	})

	jwt := auth.NewJWT(render)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwt.ServeHTTP(w, r, next)
	}))
	defer ts.Close()

	request, _ := http.NewRequest("GET", ts.URL, nil)
	authToken := createInvalidSigningMethodToken()
	request.Header.Set("Authorization", fmt.Sprintf("bearer %s", authToken))
	res, _ := ts.Client().Do(request)
	assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
}

func createInvalidSigningMethodToken() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, auth.VelocityClaims{
		Userame: "Bob",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 2).Unix(),
			Issuer:    "Velocity",
			NotBefore: time.Now().Unix(),
		},
	})
	token.Header["alg"] = "ES512"
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return tokenString
}

func TestInvalidClaimsToken(t *testing.T) {
	os.Setenv("JWT_TOKEN", "test")
	render := render.New()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, http.StatusOK, "Woo")
	})

	jwt := auth.NewJWT(render)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwt.ServeHTTP(w, r, next)
	}))
	defer ts.Close()

	request, _ := http.NewRequest("GET", ts.URL, nil)
	authToken := createInvalidClaimsToken()
	request.Header.Set("Authorization", fmt.Sprintf("bearer %s", authToken))
	res, _ := ts.Client().Do(request)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func createInvalidClaimsToken() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, auth.VelocityClaims{
		Userame: "Bob",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Unix() - 500,
			Issuer:    "Velocity",
			NotBefore: time.Now().Unix() - 1000,
		},
	})

	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return tokenString
}
