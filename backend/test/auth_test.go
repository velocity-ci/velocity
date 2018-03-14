package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"
)

func iAmAuthenticated() error {

	err := theFollowingUsersExist(&gherkin.DataTable{
		Rows: []*gherkin.TableRow{
			{
				Cells: []*gherkin.TableCell{
					{
						Value: "username",
					},
					{
						Value: "password",
					},
				},
			},
			{
				Cells: []*gherkin.TableCell{
					{
						Value: "admin",
					},
					{
						Value: "testPassword",
					},
				},
			},
		},
	})

	if err != nil {
		return err
	}

	err = iAuthenticateWithTheFollowingCredentials(&gherkin.DataTable{
		Rows: []*gherkin.TableRow{
			{
				Cells: []*gherkin.TableCell{
					{
						Value: "username",
					},
					{
						Value: "admin",
					},
				},
			},
			{
				Cells: []*gherkin.TableCell{
					{
						Value: "password",
					},
					{
						Value: "testPassword",
					},
				},
			},
		},
	})

	var r map[string]interface{}
	json.Unmarshal(responseBody, &r)

	headers.Set("Authorization", fmt.Sprintf("Bearer %s", r["token"].(string)))

	if err != nil {
		return err
	}

	return nil
}

func theFollowingUsersExist(userTable *gherkin.DataTable) error {
	uM := user.NewManager(app.DB, valid, trans)

	for _, r := range userTable.Rows[1:] {
		_, err := uM.Create(r.Cells[0].Value, r.Cells[1].Value)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func iAuthenticateWithTheFollowingCredentials(credsTable *gherkin.DataTable) error {

	username := credsTable.Rows[0].Cells[1].Value
	password := credsTable.Rows[1].Cells[1].Value

	authPayload, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/v1/auth", testServer.URL),
		bytes.NewBuffer(authPayload),
	)

	req.Header = headers

	response, _ = client.Do(req)
	responseBody, _ = ioutil.ReadAll(response.Body)
	response.Body.Close()

	return nil
}
