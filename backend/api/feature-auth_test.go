package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/velocity-ci/velocity/backend/api/domain/user"
)

func iAmAuthenticated() error {

	err := theFollowingUsersExist(&gherkin.DataTable{
		Rows: []*gherkin.TableRow{
			&gherkin.TableRow{
				Cells: []*gherkin.TableCell{
					&gherkin.TableCell{
						Value: "username",
					},
					&gherkin.TableCell{
						Value: "password",
					},
				},
			},
			&gherkin.TableRow{
				Cells: []*gherkin.TableCell{
					&gherkin.TableCell{
						Value: "admin",
					},
					&gherkin.TableCell{
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
			&gherkin.TableRow{
				Cells: []*gherkin.TableCell{
					&gherkin.TableCell{
						Value: "username",
					},
					&gherkin.TableCell{
						Value: "admin",
					},
				},
			},
			&gherkin.TableRow{
				Cells: []*gherkin.TableCell{
					&gherkin.TableCell{
						Value: "password",
					},
					&gherkin.TableCell{
						Value: "testPassword",
					},
				},
			},
		},
	})

	var r map[string]interface{}
	json.Unmarshal(responseBody, &r)

	headers.Set("Authorization", fmt.Sprintf("bearer %s", r["authToken"].(string)))

	if err != nil {
		return err
	}

	return nil
}

func theFollowingUsersExist(userTable *gherkin.DataTable) error {
	uM := user.NewManager(db)

	for _, r := range userTable.Rows[1:] {
		u := user.User{Username: r.Cells[0].Value}
		u.HashPassword(r.Cells[1].Value)

		uM.Save(u)
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

	response, _ = client.Do(req)
	responseBody, _ = ioutil.ReadAll(response.Body)
	response.Body.Close()

	return nil
}
