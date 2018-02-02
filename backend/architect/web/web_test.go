package web

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/go/canonical/json"
	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/architect/domain/persistence/db"

	"github.com/stretchr/testify/assert"
)

type webSuite struct {
	testServer    *httptest.Server
	httpClient    *http.Client
	clientHeaders http.Header
}

func (s *webSuite) SetupSuite() {
	os.Setenv("JWT_SECRET", "test")
	w := NewWeb()
	s.testServer = httptest.NewUnstartedServer(w.Server.Server.Handler)
	s.testServer.Config = w.Server.Server

	s.httpClient = &http.Client{
		Timeout: time.Second * 1,
	}

	s.clientHeaders = http.Header{}
	s.clientHeaders.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	s.testServer.Start()
}

func (s *webSuite) TearDownSuite() {
	s.testServer.Close()
}

func (s *webSuite) SetupTest() {
	// clean the database before every scenario
	truncateStmt := "TRUNCATE TABLE"
	switch db.GetDialectName() {
	case "sqlite3":
		truncateStmt = "DELETE FROM"
		break
	}

	// err := db.Exec(fmt.Sprintf(`
	// 	%[1]s build_step_streams;
	// 	%[1]s build_steps;
	// 	%[1]s builds;
	// 	%[1]s tasks;
	// 	%[1]s commit_branches;
	// 	%[1]s branches;
	// 	%[1]s commits;
	// 	%[1]s knownhosts;
	// 	%[1]s users;
	// `, truncateStmt))

	err := db.Exec(fmt.Sprintf(`
			%[1]s projects;	
			%[1]s knownhosts;
			%[1]s users;
		`, truncateStmt))

	if err != nil {
		log.Printf("could not truncate database %s", err)
	}
}

type respMap struct {
	Key      string
	Type     string
	Expected string
}

func validateResp(t *testing.T, body []byte, expected []respMap) {
	bodyMap := map[string]interface{}{}
	json.Unmarshal(body, &bodyMap)
	for _, rM := range expected {
		val, err := recurseChars(rM.Key, bodyMap)
		if err != nil {
			log.Println(err)
			t.FailNow()
		}
		rM.Assert(t, val)
	}
}

func recurseChars(eAttr string, resp map[string]interface{}) (interface{}, error) {
	// var key string
	var keyAt int
	var index int
	var indexAt int
	var indexLength int

	// check for object
	re := regexp.MustCompile(`\.\w+`)
	loc := re.FindStringIndex(eAttr)
	if loc != nil {
		// key = eAttr[loc[0]+1 : loc[1]]
		keyAt = loc[0] + 1
	}

	re = regexp.MustCompile(`\[\d+\]`)
	loc = re.FindStringIndex(eAttr)
	if loc != nil {
		index, _ = strconv.Atoi(eAttr[loc[0]+1 : loc[1]-1])
		indexAt = loc[0] + 1
		indexLength = loc[1] - loc[0]
	}

	if (keyAt > 0 && indexAt > 0) && // both key and index found
		(keyAt < indexAt) { // key before index
		return recurseChars(eAttr[keyAt:], resp[eAttr].(map[string]interface{}))
	}

	if ((keyAt > 0 && indexAt > 0) && (keyAt > indexAt)) || // both key and index found and key after index
		(keyAt == 0 && indexAt > 0) { // only index found
		parts := strings.Split(eAttr, "[")
		switch x := resp[parts[0]].(type) {
		case []interface{}:
			if (indexAt + indexLength) < len(eAttr) {
				return recurseChars(eAttr[indexAt+indexLength:], x[index].(map[string]interface{}))
			} else {
				l := map[string]interface{}{}
				for i, v := range x {
					l[strconv.Itoa(i)] = v
				}
				return recurseChars(strconv.Itoa(index), l)
			}
		default:
			return nil, fmt.Errorf("could not determine type: %s", eAttr)
		}
	}

	return resp[eAttr], nil

}

func (rM respMap) Assert(t *testing.T, v interface{}) {
	switch rM.Type {
	case "string":
		rM.assertString(t, v)
		break
	case "time":
		rM.assertTime(t, v)
		break
	case "int":
		rM.assertInt(t, v)
		break
	default:
		log.Printf("invalid type %s for %s", rM.Type, rM.Key)
		t.Fail()
	}
}

func (rM respMap) assertString(t *testing.T, val interface{}) {
	v := val.(string)
	switch rM.Expected {
	case "*any":
		assert.NotEmpty(t, v)
	default:
		assert.Equal(t, rM.Expected, v)
	}
}

func (rM respMap) assertTime(t *testing.T, val interface{}) {
	v, err := time.Parse(time.RFC3339, val.(string))
	if err != nil {
		t.Errorf("could not parse time %s", val)
	}

	switch rM.Expected[:1] {
	case "*":
		duration, err := time.ParseDuration(rM.Expected[1:])
		if err != nil {
			t.Errorf("could not parse duration %s", rM.Expected[1:])
		}
		assert.WithinDuration(t, time.Now(), v, duration)
	}

}

func (rM respMap) assertInt(t *testing.T, val interface{}) {
	v := val.(float64)
	eV, err := strconv.Atoi(rM.Expected)
	if err != nil {
		t.Errorf("could not parse int %s", rM.Expected)
	}

	assert.Equal(t, v, float64(eV))
}
