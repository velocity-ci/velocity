package project_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/master/velocity/app"
	"github.com/velocity-ci/velocity/master/velocity/domain"
)

var accessToken string
var auth *domain.UserAuth

func ServerTest(t *testing.T, f func(*testing.T, *http.Client, string)) {
	os.Setenv("DB_PATH", fmt.Sprintf("%s.db", t.Name()))
	velocity := app.New()
	defer velocity.Stop()
	s := httptest.NewServer(velocity.(*app.VelocityAPI).Router.Negroni)
	defer s.Close()
	defer os.Remove(fmt.Sprintf("%s.db", t.Name()))
	defer os.Remove("/root/.ssh/known_hosts")

	c := s.Client()

	f(t, c, s.URL)
}

func login(t *testing.T, c *http.Client, baseURL string) {
	j, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin1234"})
	resp, err := c.Post(fmt.Sprintf("%s/v1/auth", baseURL), "application/json", bytes.NewBuffer(j))
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 201, "Could not log in.")

	auth = &domain.UserAuth{}
	json.NewDecoder(resp.Body).Decode(auth)
}

func TestGetProjects(t *testing.T) {
	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
		login(t, c, baseURL)

		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/projects", baseURL), nil)
		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", auth.Token))

		resp, err := c.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, resp.StatusCode, 200, "Could not retrieve projects.")

		defer resp.Body.Close()
		respProjects := []domain.ResponseProject{}
		json.NewDecoder(resp.Body).Decode(&respProjects)

		assert.Equal(t, len(respProjects), 0, "Too many projects returned.")
	})
}

func TestUnauthenticatedGetProjects(t *testing.T) {
	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/projects", baseURL), nil)

		resp, err := c.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, resp.StatusCode, 401, "Unauthenticated access allowed.")
	})
}

func TestPostProjectsHTTPSRepo(t *testing.T) {
	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
		login(t, c, baseURL)

		j, _ := json.Marshal(map[string]string{
			"name":       "velocity",
			"repository": "https://github.com/velocity-ci/velocity.git",
		})

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/projects", baseURL), bytes.NewBuffer(j))
		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", auth.Token))

		resp, err := c.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, resp.StatusCode, 201, "Could not create HTTP Project")

		req, _ = http.NewRequest("GET", fmt.Sprintf("%s/v1/projects", baseURL), nil)
		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", auth.Token))

		resp, err = c.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, resp.StatusCode, 200, "Could not retrieve projects.")

		defer resp.Body.Close()
		respProjects := []domain.ResponseProject{}
		json.NewDecoder(resp.Body).Decode(&respProjects)

		assert.Equal(t, respProjects[0].ID, "velocity")
		assert.Equal(t, respProjects[0].Name, "velocity")
		assert.Equal(t, respProjects[0].Repository, "https://github.com/velocity-ci/velocity.git")
		assert.Equal(t, respProjects[0].Synchronising, false)
	})
}

func TestPostProjectsGITRepo(t *testing.T) {
	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
		login(t, c, baseURL)

		j, _ := json.Marshal(map[string]string{
			"entry": "github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==",
		})

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/ssh/known-hosts", baseURL), bytes.NewBuffer(j))
		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", auth.Token))

		resp, err := c.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, resp.StatusCode, 201, "Could not create known host for github.com")

		j, _ = json.Marshal(map[string]string{
			"name":       "velocity",
			"repository": "git@github.com:velocity-ci/velocity.git",
			"key": `-----BEGIN RSA PRIVATE KEY-----
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
70/zmQKCAQAc4323mF1IRKeGFLTeu0DbV8Jv41TziJUfL8L+RvUNq4yu0M9Mwrl4
o/jr9tiQdlZrQsgq10PTc+EJ8Ex+R2ea2ZpxiT9JTtNeJe+IysqRsmR+2dFqbGB2
yIpeV0CTf6hDeocax9DsB9/HtR2T4uYjv2QRakwojWs5zJqU0mC29+j/SZum+Xcw
F8oz5uJJ8U42fNSLYz1iQ6VrK8AK70YYilGG3ssG9wxeYH5lsTPAH3+4q0n5HcsM
hNyeXuZlZrth6t2sFlWYJfisXk8wG9v04ibvg8xrIRS/Y2SdYrobqisidXXuu3rp
GxGHecfFH3C5LCmv37RHEXFyVYbpfQ/BAoIBABetjvjx25by37I0J1hfuT4Hx23F
Xff8WCRN2Zh/S0o5RyDzmueMdLjyqYfpZD61txwVCHm5wdPoEEPT9M6dSE1Yjnmw
P1N6hhzpXGIs5hR21o8EcYtzWRshKmbo4oAIwAe6L3E1NB1zY/ccqNdyXZmwUwa1
Efje6IstnCE7dLSrkqBSJG1KUT6Z+aNzP9iugy/5QLhj4SBVx2Rhzn4wwud+p4fy
FHJ6a6vM0DFVf9YqmE1El2RB11Pe8s5C4XCqbYDnZt23OQhEApYccS6FUxc0Equi
hQG53FH+t3G8yeJhy2drCYjl9mO52fuLfApaalb/X6oW7+fQyh9nasr8TnE=
-----END RSA PRIVATE KEY-----`,
		})

		req, _ = http.NewRequest("POST", fmt.Sprintf("%s/v1/projects", baseURL), bytes.NewBuffer(j))
		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", auth.Token))

		resp, err = c.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, resp.StatusCode, 201, "Could not create GIT Project")

		req, _ = http.NewRequest("GET", fmt.Sprintf("%s/v1/projects", baseURL), nil)
		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", auth.Token))

		resp, err = c.Do(req)
		assert.Nil(t, err)
		assert.Equal(t, resp.StatusCode, 200, "Could not retrieve projects.")

		defer resp.Body.Close()
		respProjects := []domain.ResponseProject{}
		json.NewDecoder(resp.Body).Decode(&respProjects)

		assert.Equal(t, respProjects[0].ID, "velocity")
		assert.Equal(t, respProjects[0].Name, "velocity")
		assert.Equal(t, respProjects[0].Repository, "git@github.com:velocity-ci/velocity.git")
		assert.Equal(t, respProjects[0].Synchronising, false)
	})
}
