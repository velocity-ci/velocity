package web

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/docker/go/canonical/json"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/velocity-ci/velocity/backend/architect/domain"
	"github.com/velocity-ci/velocity/backend/architect/domain/persistence"
)

func GivenIAmAuthenticated(web *webSuite) {
	// Given the admin user exists with password
	u, _ := domain.NewUser("admin", "test12345")
	persistence.SaveUser(u)

	// When I send a request to /v1/auth with body
	reqAuthJSON := `
	{
		"username": "admin",
		"password": "test12345"
	}`

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s%s", web.testServer.URL, "/v1/auth"),
		bytes.NewBuffer([]byte(reqAuthJSON)),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	response, _ := web.httpClient.Do(req)
	responseBody, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	var r map[string]interface{}
	json.Unmarshal(responseBody, &r)

	web.clientHeaders.Set("Authorization", fmt.Sprintf("Bearer %s", r["token"].(string)))
}

type AuthSuite struct {
	suite.Suite
	webSuite
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}

func (s *AuthSuite) TestValidAuthUser() {
	// Given the admin user exists with password
	u, _ := domain.NewUser("admin", "test12345")
	persistence.SaveUser(u)

	// When I send a request to /v1/auth with body
	reqAuthJSON := `
	{
		"username": "admin",
		"password": "test12345"
	}`

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s%s", s.testServer.URL, "/v1/auth"),
		bytes.NewBuffer([]byte(reqAuthJSON)),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	response, _ := s.httpClient.Do(req)
	responseBody, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	// Then
	assert.Equal(s.T(), http.StatusCreated, response.StatusCode)
	validateResp(s.T(), responseBody, []respMap{
		respMap{Key: "username", Type: "string", Expected: "admin"},
		respMap{Key: "token", Type: "string", Expected: "*any"},
		respMap{Key: "expires", Type: "time", Expected: "*+48h"},
	})
}

func (s *AuthSuite) TestBadPayloadAuthUser() {
	// Given the admin user exists with password
	u, _ := domain.NewUser("admin", "test12345")
	persistence.SaveUser(u)

	// When I send a request to /v1/auth with body
	reqAuthJSON := `
	{
		"username": "admin"
		"password": "test12345"
	}`

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s%s", s.testServer.URL, "/v1/auth"),
		bytes.NewBuffer([]byte(reqAuthJSON)),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	response, _ := s.httpClient.Do(req)
	// responseBody, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	// Then
	assert.Equal(s.T(), http.StatusBadRequest, response.StatusCode)
}

func (s *AuthSuite) TestInvalidAuthUser() {
	// Given the admin user exists with password
	u, _ := domain.NewUser("admin", "test12345")
	persistence.SaveUser(u)

	// When I send a request to /v1/auth with body
	reqAuthJSON := `
	{
		"username": "ad",
		"password": "test12345"
	}`

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s%s", s.testServer.URL, "/v1/auth"),
		bytes.NewBuffer([]byte(reqAuthJSON)),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	response, _ := s.httpClient.Do(req)
	responseBody, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	// Then
	assert.Equal(s.T(), http.StatusBadRequest, response.StatusCode)
	validateResp(s.T(), responseBody, []respMap{
		respMap{Key: "username[0]", Type: "string", Expected: "username must be at least 3 characters in length"},
	})
}

func (s *AuthSuite) TestInvalidPasswordAuthUser() {
	// Given the admin user exists with password
	u, _ := domain.NewUser("admin", "test12345")
	persistence.SaveUser(u)

	// When I send a request to /v1/auth with body
	reqAuthJSON := `
	{
		"username": "admin",
		"password": "test123456"
	}`

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s%s", s.testServer.URL, "/v1/auth"),
		bytes.NewBuffer([]byte(reqAuthJSON)),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	response, _ := s.httpClient.Do(req)
	response.Body.Close()

	// Then
	assert.Equal(s.T(), http.StatusUnauthorized, response.StatusCode)
}

func (s *AuthSuite) TestInvalidUsernameAuthUser() {
	// Given the admin user exists with password
	u, _ := domain.NewUser("admin", "test12345")
	persistence.SaveUser(u)

	// When I send a request to /v1/auth with body
	reqAuthJSON := `
	{
		"username": "admina",
		"password": "test12345"
	}`

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s%s", s.testServer.URL, "/v1/auth"),
		bytes.NewBuffer([]byte(reqAuthJSON)),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	response, _ := s.httpClient.Do(req)
	response.Body.Close()

	// Then
	assert.Equal(s.T(), http.StatusUnauthorized, response.StatusCode)
}
