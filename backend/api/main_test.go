package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/jinzhu/gorm"
)

var testServer *httptest.Server
var app *VelocityAPI
var db *gorm.DB

var client *http.Client
var response *http.Response
var responseBody []byte

func FeatureContext(s *godog.Suite) {
	s.Step(`^the following users exist:$`, theFollowingUsersExist)
	s.Step(`^I authenticate with the following credentials:$`, iAuthenticateWithTheFollowingCredentials)
	s.Step(`^the response has status "([^"]*)"$`, theResponseHasStatus)
	s.Step(`^the response has the following attributes:$`, theResponseHasTheFollowingAttributes)

	s.BeforeSuite(func() {
		db = NewGORMDB("test.db")
		app = NewVelocity(db)
		testServer = httptest.NewUnstartedServer(app.Router.Negroni)
		testServer.Config = app.server

		client = &http.Client{
			Timeout: time.Second * 10,
		}

		testServer.Start()
	})

	s.BeforeScenario(func(interface{}) {
		// clean the database before every scenario
		truncateStmt := "TRUNCATE TABLE"
		switch db.Dialect().GetName() {
		case "sqlite3":
			truncateStmt = "DELETE FROM"
			break
		}

		_, err := db.DB().Exec(fmt.Sprintf(`
			%[1]s build_step_streams;
			%[1]s build_steps;
			%[1]s builds;
			%[1]s tasks;
			%[1]s commit_branches;
			%[1]s branches;
			%[1]s commits;
			%[1]s projects;
			%[1]s knownhosts;
			%[1]s users;
		`, truncateStmt))

		if err != nil {
			log.Printf("could not truncate database %s", err)
		}
	})

	s.AfterScenario(func(interface{}, error) {
	})

	s.AfterSuite(func() {
		app.Stop()
		testServer.Close()
		db.Close()
		os.Remove("test.db")
	})
}
