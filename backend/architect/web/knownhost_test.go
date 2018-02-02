package web

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/velocity-ci/velocity/backend/architect/domain"
	"github.com/velocity-ci/velocity/backend/architect/domain/persistence"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type KnownHostSuite struct {
	suite.Suite
	webSuite
}

func TestKnownHostSuite(t *testing.T) {
	suite.Run(t, new(KnownHostSuite))
}

// Background
func (s *KnownHostSuite) SetupTest() {
	s.webSuite.SetupTest()

	// Given I am authenticated
	GivenIAmAuthenticated(&s.webSuite)
}

func (s *KnownHostSuite) TestCreateValidKnownHost() {

	// When I send a request
	reqKnownHostJSON := `
	{
		"entry": "github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ=="
	}`

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s%s", s.testServer.URL, "/v1/ssh/known-hosts"),
		bytes.NewBuffer([]byte(reqKnownHostJSON)),
	)
	req.Header = s.webSuite.clientHeaders

	response, _ := s.httpClient.Do(req)
	responseBody, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	// Then
	assert.Equal(s.T(), http.StatusCreated, response.StatusCode)
	validateResp(s.T(), responseBody, []respMap{
		respMap{Key: "id", Type: "string", Expected: "*any"},
		respMap{Key: "hosts[0]", Type: "string", Expected: "github.com"},
		respMap{Key: "comment", Type: "string", Expected: ""},
		respMap{Key: "sha256", Type: "string", Expected: "SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8"},
		respMap{Key: "md5", Type: "string", Expected: "16:27:ac:a5:76:28:2d:36:63:1b:56:4d:eb:df:a6:48"},
	})
}

func (s *KnownHostSuite) TestCreateInvalidKnownHost() {

	// When I send a request
	reqKnownHostJSON := `
	{
		"entry": "github.com ssh-rsa AAAAB3NzaC1yc2AAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ=="
	}`

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s%s", s.testServer.URL, "/v1/ssh/known-hosts"),
		bytes.NewBuffer([]byte(reqKnownHostJSON)),
	)
	req.Header = s.webSuite.clientHeaders

	response, _ := s.httpClient.Do(req)
	responseBody, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	// Then
	assert.Equal(s.T(), http.StatusBadRequest, response.StatusCode)
	validateResp(s.T(), responseBody, []respMap{
		respMap{Key: "entry[0]", Type: "string", Expected: "entry is not a valid key!"},
	})
}

func (s *KnownHostSuite) TestCreateBadPayloadKnownHost() {

	// When I send a request
	reqKnownHostJSON := `
	{
		"entry": "github.com ssh-rsa AAAAB3NzaC1yc2AAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==",
	}`

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s%s", s.testServer.URL, "/v1/ssh/known-hosts"),
		bytes.NewBuffer([]byte(reqKnownHostJSON)),
	)
	req.Header = s.webSuite.clientHeaders

	response, _ := s.httpClient.Do(req)
	response.Body.Close()

	// Then
	assert.Equal(s.T(), http.StatusBadRequest, response.StatusCode)
}

func (s *KnownHostSuite) TestListKnownHosts() {
	// Given the following known hosts exist
	k, _ := domain.NewKnownHost("github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==")
	persistence.SaveKnownHost(k)

	// When I send a request to list
	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("%s%s", s.testServer.URL, "/v1/ssh/known-hosts"),
		nil,
	)
	req.Header = s.webSuite.clientHeaders
	response, _ := s.httpClient.Do(req)
	responseBody, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()

	// Then
	validateResp(s.T(), responseBody, []respMap{
		respMap{Key: "total", Type: "int", Expected: "1"},
		respMap{Key: "data[0].id", Type: "string", Expected: "*any"},
		respMap{Key: "data[0].hosts[0]", Type: "string", Expected: "github.com"},
		respMap{Key: "data[0].sha256", Type: "string", Expected: "SHA256:nThbg6kXUpJWGl7E1IGOCspRomTxdCARLviKw6E5SY8"},
		respMap{Key: "data[0].md5", Type: "string", Expected: "16:27:ac:a5:76:28:2d:36:63:1b:56:4d:eb:df:a6:48"},
	})
}
