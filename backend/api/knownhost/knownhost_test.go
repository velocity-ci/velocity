package knownhost_test

// var accessToken string
// var auth *UserAuth

// func ServerTest(t *testing.T, f func(*testing.T, *http.Client, string)) {
// 	os.Setenv("DB_PATH", fmt.Sprintf("%s.db", t.Name()))
// 	velocity := app.New()
// 	defer velocity.Stop()
// 	s := httptest.NewServer(velocity.(*app.VelocityAPI).Router.Negroni)
// 	defer s.Close()
// 	defer os.Remove(fmt.Sprintf("%s.db", t.Name()))
// 	defer os.Remove("/root/.ssh/known_hosts")

// 	c := s.Client()

// 	f(t, c, s.URL)
// }

// func login(c *http.Client, baseURL string) {
// 	j, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin1234"})
// 	resp, err := c.Post(fmt.Sprintf("%s/v1/auth", baseURL), "application/json", bytes.NewBuffer(j))
// 	if err != nil {
// 		log.Fatal(err)
// 		return
// 	}
// 	if resp.StatusCode != 201 {
// 		log.Fatal("Could not log in")
// 		return
// 	}

// 	auth = &UserAuth{}
// 	json.NewDecoder(resp.Body).Decode(auth)
// }

// func TestGetKnownHosts(t *testing.T) {
// 	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
// 		login(c, baseURL)

// 		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/ssh/known-hosts", baseURL), nil)
// 		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", auth.Token))

// 		resp, err := c.Do(req)
// 		if err != nil {
// 			log.Fatal(err)
// 			t.Fail()
// 		}
// 		if resp.StatusCode != 200 {
// 			t.Fail()
// 		}
// 	})
// }

// func TestUnauthenticatedGetKnownHosts(t *testing.T) {
// 	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
// 		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/ssh/known-hosts", baseURL), nil)

// 		resp, err := c.Do(req)
// 		if err != nil {
// 			log.Fatal(err)
// 			t.Fail()
// 		}
// 		if resp.StatusCode != 401 {
// 			t.Fail()
// 		}
// 	})
// }

// func TestPostKnownHosts(t *testing.T) {
// 	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
// 		login(c, baseURL)

// 		j, _ := json.Marshal(map[string]string{
// 			"entry": "github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==",
// 		})

// 		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/ssh/known-hosts", baseURL), bytes.NewBuffer(j))
// 		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", auth.Token))

// 		resp, err := c.Do(req)
// 		if err != nil {
// 			log.Fatal(err)
// 			t.Fail()
// 		}
// 		if resp.StatusCode != 201 {
// 			t.Fail()
// 		}

// 		req, _ = http.NewRequest("GET", fmt.Sprintf("%s/v1/ssh/known-hosts", baseURL), nil)
// 		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", auth.Token))

// 		resp, err = c.Do(req)
// 		if err != nil {
// 			log.Fatal(err)
// 			t.Fail()
// 		}
// 		if resp.StatusCode != 200 {
// 			t.Fail()
// 		}
// 	})
// }

// func TestPostInvalidKnownHosts(t *testing.T) {
// 	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
// 		login(c, baseURL)

// 		j, _ := json.Marshal(map[string]string{
// 			"entry": "github.com ssh-rsa AAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==",
// 		})

// 		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/ssh/known-hosts", baseURL), bytes.NewBuffer(j))
// 		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", auth.Token))

// 		resp, err := c.Do(req)
// 		if err != nil {
// 			log.Fatal(err)
// 			t.Fail()
// 		}
// 		if resp.StatusCode != 400 {
// 			t.Fail()
// 		}
// 	})
// }

// func TestPostInvalidJSON(t *testing.T) {
// 	ServerTest(t, func(t *testing.T, c *http.Client, baseURL string) {
// 		login(c, baseURL)

// 		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/v1/ssh/known-hosts", baseURL), bytes.NewBuffer([]byte("aaa")))
// 		req.Header.Set("Authorization", fmt.Sprintf("bearer %s", auth.Token))

// 		resp, err := c.Do(req)
// 		if err != nil {
// 			log.Fatal(err)
// 			t.Fail()
// 		}
// 		if resp.StatusCode != 400 {
// 			t.Fail()
// 		}
// 	})
// }
