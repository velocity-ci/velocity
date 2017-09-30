package auth_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/velocity-ci/velocity/master/velocity/app"
)

func ServerTest(t *testing.T, f func(*testing.T, *http.Client, string)) {
	os.Setenv("DB_PATH", fmt.Sprintf("%s.db", t.Name()))
	velocity := app.New()
	defer velocity.Stop()
	s := httptest.NewServer(velocity.(*app.VelocityAPI).Router.Negroni)
	defer s.Close()
	defer os.Remove(fmt.Sprintf("%s.db", t.Name()))

	c := s.Client()

	f(t, c, s.URL)
}

func TestInvalidCredentials(t *testing.T) {
	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
		j, _ := json.Marshal(map[string]string{"username": "admin", "password": "123456789"})
		resp, err := c.Post(fmt.Sprintf("%s/v1/auth", baseURL), "application/json", bytes.NewBuffer(j))
		if err != nil {
			log.Fatal(err)
			t.Fail()
		}
		if resp.StatusCode != 401 {
			t.Fail()
		}
	})
}

func TestInvalidJSON(t *testing.T) {
	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
		resp, err := c.Post(fmt.Sprintf("%s/v1/auth", baseURL), "application/json", bytes.NewBuffer([]byte("aaaaa")))
		if err != nil {
			log.Fatal(err)
			t.Fail()
		}
		if resp.StatusCode != 400 {
			t.Fail()
		}
	})
}

func TestInvalidUsername(t *testing.T) {
	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
		j, _ := json.Marshal(map[string]string{"username": "ad", "password": "123456789"})
		resp, err := c.Post(fmt.Sprintf("%s/v1/auth", baseURL), "application/json", bytes.NewBuffer(j))
		if err != nil {
			log.Fatal(err)
			t.Fail()
		}
		if resp.StatusCode != 400 {
			t.Fail()
		}
	})
}

func TestInvalidPassword(t *testing.T) {
	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
		j, _ := json.Marshal(map[string]string{"username": "admin", "password": "123"})
		resp, err := c.Post(fmt.Sprintf("%s/v1/auth", baseURL), "application/json", bytes.NewBuffer(j))
		if err != nil {
			log.Fatal(err)
			t.Fail()
		}
		if resp.StatusCode != 400 {
			t.Fail()
		}
	})
}

func TestMissingUsername(t *testing.T) {
	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
		j, _ := json.Marshal(map[string]string{"username": "adminqwe", "password": "12asdasd3"})
		resp, err := c.Post(fmt.Sprintf("%s/v1/auth", baseURL), "application/json", bytes.NewBuffer(j))
		if err != nil {
			log.Fatal(err)
			t.Fail()
		}
		if resp.StatusCode != 401 {
			t.Fail()
		}
	})
}

func TestValidCredentials(t *testing.T) {
	f := func(t *testing.T, c *http.Client, baseURL string) {
		j, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin1234"})
		resp, err := c.Post(fmt.Sprintf("%s/v1/auth", baseURL), "application/json", bytes.NewBuffer(j))
		if err != nil {
			log.Fatal(err)
			t.Fail()
		}
		if resp.StatusCode != 201 {
			t.Fail()
		}
	}
	ServerTest(t, f)
}
