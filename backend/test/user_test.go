package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/docker/go/canonical/json"
)

func iCreateTheFollowingUser(userTable *gherkin.DataTable) error {
	entryPayload, _ := json.Marshal(map[string]string{
		"username": userTable.Rows[0].Cells[1].Value,
		"password": userTable.Rows[1].Cells[1].Value,
	})

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/v1/users", testServer.URL),
		bytes.NewBuffer(entryPayload),
	)

	req.Header = headers

	response, _ = client.Do(req)
	responseBody, _ = ioutil.ReadAll(response.Body)
	response.Body.Close()

	return nil
}

func iListTheUsers() error {
	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/v1/users", testServer.URL),
		nil,
	)

	req.Header = headers

	response, _ = client.Do(req)
	responseBody, _ = ioutil.ReadAll(response.Body)
	response.Body.Close()

	return nil
}
