package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/docker/go/canonical/json"
)

func iCreateTheFollowingKnownHost(entry *gherkin.DocString) error {
	entryPayload, _ := json.Marshal(map[string]string{
		"entry": entry.Content,
	})

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/v1/ssh/known-hosts", testServer.URL),
		bytes.NewBuffer(entryPayload),
	)

	req.Header = headers

	response, _ = client.Do(req)
	responseBody, _ = ioutil.ReadAll(response.Body)
	response.Body.Close()

	return nil
}

func theFollowingKnownHostExists(entry *gherkin.DocString) error {
	return iCreateTheFollowingKnownHost(entry)
}

func iListTheKnownHosts() error {
	req, _ := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/v1/ssh/known-hosts", testServer.URL),
		nil,
	)

	req.Header = headers

	response, _ = client.Do(req)
	responseBody, _ = ioutil.ReadAll(response.Body)
	response.Body.Close()

	return nil
}
