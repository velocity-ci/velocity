package build

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/docker/go/canonical/json"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
)

type Resolver struct {
	// BuildValidator *BuildValidator
	commitManager commit.Repository
}

func NewResolver(commitManager *commit.Manager) *Resolver {
	return &Resolver{
		commitManager: commitManager,
		// BuildValidator: buildValidator,
	}
}

func (r *Resolver) BuildFromRequest(b io.ReadCloser, t task.Task) (Build, error) {
	reqBuild := RequestBuild{}

	err := json.NewDecoder(b).Decode(&reqBuild)
	if err != nil {
		return Build{}, err
	}

	for i, rP := range reqBuild.Parameters {
		reqBuild.Parameters[i].Value = strings.TrimSpace(rP.Value)
	}

	// err = r.buildValidator.Validate(&reqBuild) Should validate params match up

	// if err != nil {
	// 	return nil, err
	// }

	cm, err := r.commitManager.GetCommitByCommitID(t.CommitID)
	if err != nil {
		log.Printf("could not find commit %s?!?!", t.CommitID)
	}

	build := NewBuild(cm.ProjectID, t, getParametersFromRequest(reqBuild.Parameters))

	return build, nil
}

func BuildQueryOptsFromRequest(r *http.Request) BuildQuery {
	reqQueries := r.URL.Query()

	amount := uint64(15)
	if a, err := strconv.ParseUint(reqQueries.Get("amount"), 10, 64); err == nil {
		amount = a
	}

	page := uint64(1)
	if p, err := strconv.ParseUint(reqQueries.Get("page"), 10, 64); err == nil {
		page = p
	}

	status := "all"
	if len(reqQueries.Get("status")) > 1 {
		status = reqQueries.Get("status")
	}

	return BuildQuery{
		Status: status,
		Amount: amount,
		Page:   page,
	}
}

func StreamLineQueryOptsFromRequest(r *http.Request) StreamLineQuery {
	reqQueries := r.URL.Query()

	amount := uint64(200)
	if a, err := strconv.ParseUint(reqQueries.Get("amount"), 10, 64); err == nil {
		amount = a
	}

	page := uint64(1)
	if p, err := strconv.ParseUint(reqQueries.Get("page"), 10, 64); err == nil {
		page = p
	}

	return StreamLineQuery{
		Amount: amount,
		Page:   page,
	}
}

func getParametersFromRequest(reqParams []RequestParameter) map[string]string {
	res := map[string]string{}
	for _, reqParam := range reqParams {
		res[reqParam.Name] = reqParam.Value
	}

	return res
}
