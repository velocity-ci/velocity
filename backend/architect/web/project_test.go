package web

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProjectSuite struct {
	suite.Suite
	webSuite
}

func TestProjectSuite(t *testing.T) {
	suite.Run(t, new(ProjectSuite))
}

// Background
func (s *ProjectSuite) SetupTest() {
	s.webSuite.SetupTest()

	// Given I am authenticated
	GivenIAmAuthenticated(&s.webSuite)
}

func (s *ProjectSuite) TestValidHTTPProject() {

	// When I send a request to /v1/projects with body
	reqAuthJSON := `
	{
		"name": "Velocity http",
		"repositoryAddress": "http://localhost:3000/velocity/velocity_public.git"
	}`

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s%s", s.testServer.URL, "/v1/projects"),
		bytes.NewBuffer([]byte(reqAuthJSON)),
	)
	req.Header = s.webSuite.clientHeaders

	response, _ := s.httpClient.Do(req)
	responseBody, _ := ioutil.ReadAll(response.Body)
	log.Println(responseBody)
	response.Body.Close()

	// Then
	assert.Equal(s.T(), http.StatusCreated, response.StatusCode)
	// validateResp(s.T(), responseBody, []respMap{
	// 	respMap{Key: "username", Type: "string", Expected: "admin"},
	// 	respMap{Key: "token", Type: "string", Expected: "*any"},
	// 	respMap{Key: "expires", Type: "time", Expected: "*+48h"},
	// })
}
