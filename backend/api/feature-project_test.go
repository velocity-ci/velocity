package main

import (
	"github.com/DATA-DOG/godog/gherkin"
)

func iCreateTheFollowingProject(reqAttrs *gherkin.DataTable) error {

	sendPOSTWithAttrsTable(reqAttrs, "/v1/projects")
	return nil
}
