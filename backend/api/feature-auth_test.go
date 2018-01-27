package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/DATA-DOG/godog/gherkin"
	"github.com/velocity-ci/velocity/backend/api/domain/user"
)

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

func theResponseHasStatus(expectedStatus string) error {
	if response.Status == expectedStatus {
		return nil
	}

	return fmt.Errorf("expected: %s, got: %s", expectedStatus, response.Status)
}

func theResponseHasTheFollowingAttributes(expectedAttrs *gherkin.DataTable) error {

	resp := map[string]interface{}{}
	err := json.Unmarshal(responseBody, &resp)
	if err != nil {
		return err
	}

	for _, r := range expectedAttrs.Rows[1:] {
		eAttr := r.Cells[0].Value
		eType := r.Cells[1].Value
		eVal := r.Cells[2].Value
		dAttr := resp[eAttr]

		err := compareAttrs(dAttr, eAttr, eType, eVal)
		if err != nil {
			return err
		}
	}

	return nil
}

func compareAttrs(val interface{}, eAttr, eType, eVal string) error {
	var dVal interface{}
	var compareTypeFunc func(string, string, interface{}) error
	switch eType {
	case "string":
		dVal = ""
		compareTypeFunc = compareStringType
		break
	case "timestamp":
		dVal = time.Time{}
		compareTypeFunc = compareTimestampType
		break
	default:
		return fmt.Errorf("invalid type %s", eAttr)
		break
	}

	dJSONAttr, _ := json.Marshal(val)
	err := json.Unmarshal(dJSONAttr, &dVal)
	if err != nil {
		return err
	}

	return compareTypeFunc(eAttr, eVal, dVal)
}

func compareStringType(eAttr, eVal string, dVal interface{}) error {
	v := dVal.(string)

	switch eVal {
	case "*any":
		if len(v) < 1 {
			return fmt.Errorf("%s was empty", eAttr)
		}
	default:
		if eVal != v {
			return fmt.Errorf("expected %s, got %s", eAttr, eVal)
		}
	}

	return nil
}

func compareTimestampType(eAttr, eVal string, dVal interface{}) error {
	v, err := time.Parse(time.RFC3339, dVal.(string))

	if err != nil {
		return err
	}

	switch eVal[:1] {
	case "*":
		duration, err := time.ParseDuration(eVal[1:])
		if err != nil {
			return err
		}
		e := time.Now().Add(duration)
		if v.Format(time.ANSIC) != e.Format(time.ANSIC) {
			return fmt.Errorf("expected %s, got %s", e, v)
		}
	default:
		if eVal != dVal.(string) {
			return fmt.Errorf("expected %s, got %s", eVal, dVal)
		}
	}

	return nil
}
