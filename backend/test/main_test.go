package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/universal-translator"
	"github.com/labstack/echo"

	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"

	"github.com/DATA-DOG/godog"
	"github.com/velocity-ci/velocity/backend/pkg/architect"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	validator "gopkg.in/go-playground/validator.v9"
)

var testServer *httptest.Server
var app *architect.Architect
var valid *validator.Validate
var trans ut.Translator

var client *http.Client
var response *http.Response
var responseBody []byte
var headers http.Header

func TestMain(m *testing.M) {
	format := "progress" // non verbose mode
	concurrency := 1

	var specific bool
	for _, arg := range os.Args[1:] {
		if arg == "-test.v=true" { // go test transforms -v option - verbose mode
			format = "pretty"
			concurrency = 1
			break
		}
		if strings.Index(arg, "-test.run") == 0 {
			specific = true
		}
	}
	var status int
	if !specific {
		status = godog.RunWithOptions("godog", func(s *godog.Suite) {
			FeatureContext(s)
		}, godog.Options{
			Format:      format, // pretty format for verbose mode, otherwise - progress
			Paths:       []string{"features"},
			Concurrency: concurrency,           // concurrency for verbose mode is 1
			Randomize:   time.Now().UnixNano(), // randomize scenario execution order
		})
	}

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func FeatureContext(s *godog.Suite) {
	s.Step(`^the following users exist:$`, theFollowingUsersExist)
	s.Step(`^I authenticate with the following credentials:$`, iAuthenticateWithTheFollowingCredentials)
	s.Step(`^the response has status "([^"]*)"$`, theResponseHasStatus)
	s.Step(`^the response has the following attributes:$`, theResponseHasTheFollowingAttributes)
	s.Step(`^I am authenticated$`, iAmAuthenticated)
	s.Step(`^I create the following known host:$`, iCreateTheFollowingKnownHost)
	s.Step(`^the following known host exists:$`, theFollowingKnownHostExists)
	s.Step(`^I list the known hosts$`, iListTheKnownHosts)
	s.Step(`^I create the following user:$`, iCreateTheFollowingUser)
	s.Step(`^I list the users$`, iListTheUsers)
	// s.Step(`^I create the following project:$`, iCreateTheFollowingProject)

	s.BeforeSuite(func() {
		valid, trans = domain.NewValidator()
		app = architect.New()
		app.LogsPath = "logs/"
		app.DB = domain.NewStormDB("test.db")
		testServer = httptest.NewUnstartedServer(app.Server.Server.Handler)
		app.Init()
		testServer.Config = app.Server.Server

		client = &http.Client{
			Timeout: time.Second * 10,
		}

		headers = http.Header{}
		headers.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

		testServer.Start()
	})

	s.BeforeScenario(func(interface{}) {
		// clean the database before every scenario
		// truncateStmt := "TRUNCATE TABLE"
		// switch db.Dialect().GetName() {
		// case "sqlite3":
		// 	truncateStmt = "DELETE FROM"
		// 	break
		// }

		// _, err := db.DB().Exec(fmt.Sprintf(`
		// 	%[1]s build_step_streams;
		// 	%[1]s build_steps;
		// 	%[1]s builds;
		// 	%[1]s tasks;
		// 	%[1]s commit_branches;
		// 	%[1]s branches;
		// 	%[1]s commits;
		// 	%[1]s projects;
		// 	%[1]s knownhosts;
		// 	%[1]s users;
		// `, truncateStmt))

		// if err != nil {
		// 	log.Printf("could not truncate database %s", err)
		// }

		app.DB.Drop(&user.StormUser{})
		app.DB.Drop(&task.StormTask{})
		app.DB.Drop(&project.StormProject{})
		app.DB.Drop(&knownhost.StormKnownHost{})
		app.DB.Drop(&githistory.StormCommit{})
		app.DB.Drop(&githistory.StormBranch{})
		app.DB.Drop(&build.StormBuild{})
		app.DB.Drop(&build.StormStep{})
		app.DB.Drop(&build.StormStream{})

		app.DB.Init(&user.StormUser{})
		app.DB.Init(&task.StormTask{})
		app.DB.Init(&project.StormProject{})
		app.DB.Init(&knownhost.StormKnownHost{})
		app.DB.Init(&githistory.StormCommit{})
		app.DB.Init(&githistory.StormBranch{})
		app.DB.Init(&build.StormBuild{})
		app.DB.Init(&build.StormStep{})
		app.DB.Init(&build.StormStream{})

		// clean known hosts
		// fM := knownhost.NewFileManager()
		// fM.Clear()
	})

	s.AfterScenario(func(interface{}, error) {
	})

	s.AfterSuite(func() {
		app.Stop()
		testServer.Close()
		app.DB.Close()
		os.Remove("test.db")
	})
}
