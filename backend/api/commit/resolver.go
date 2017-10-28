package commit

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/velocity-ci/velocity/backend/api/project"
	"github.com/velocity-ci/velocity/backend/task"
)

func NewResolver(commitManager *Manager) *Resolver {
	return &Resolver{
		CommitManager: commitManager,
		// BuildValidator: buildValidator,
	}
}

type Resolver struct {
	CommitManager *Manager
	// BuildValidator *BuildValidator
}

func (r *Resolver) QueryOptsFromRequest(r *http.Request) *CommitQueryOpts {
	reqQueries := r.URL.Query()

	amount := 15
	if a, err := strconv.Atoi(reqQueries.Get("amount")); err == nil {
		amount = a
	}
	page := 1

	if p, err := strconv.Atoi(reqQueries.Get("page")); err == nil {
		page = p
	}

	return &CommitQueryOpts{
		Branch: reqQueries.Get("branch"),
		Amount: amount,
		Page:   page,
	}
}

func (r *Resolver) BuildFromRequest(b io.ReadCloser, p *project.Project, c *Commit) (*Build, error) {
	reqBuild := RequestBuild{}

	err := json.NewDecoder(b).Decode(&reqBuild)
	if err != nil {
		return nil, err
	}

	reqBuild.TaskName = strings.TrimSpace(reqBuild.TaskName)
	for i, rP := range reqBuild.Parameters {
		reqBuild.Parameters[i].Value = strings.TrimSpace(rP.Value)
	}

	// err = r.buildValidator.Validate(&reqBuild)

	// if err != nil {
	// 	return nil, err
	// }

	task, err := r.CommitManager.GetTaskForCommitInProject(c, p, reqBuild.TaskName)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	setTaskParametersFromRequest(task, reqBuild.Parameters)

	build := NewBuild(p.ID, c.Hash, task)

	return &build, nil
}

func setTaskParametersFromRequest(t *task.Task, reqParams []RequestParameter) {
	for _, reqParam := range reqParams {
		if param, ok := t.Parameters[reqParam.Name]; ok {
			param.Value = reqParam.Value
			t.Parameters[reqParam.Name] = param
		}
	}
}
