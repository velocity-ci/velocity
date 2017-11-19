package commit

import (
	"net/http"
	"strconv"
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

func QueryOptsFromRequest(r *http.Request) Query {
	reqQueries := r.URL.Query()

	amount := uint64(15)
	if a, err := strconv.ParseUint(reqQueries.Get("amount"), 10, 64); err == nil {
		amount = a
	}

	page := uint64(1)
	if p, err := strconv.ParseUint(reqQueries.Get("page"), 10, 64); err == nil {
		page = p
	}

	return Query{
		Branch: reqQueries.Get("branch"),
		Amount: amount,
		Page:   page,
	}
}

// func (r *Resolver) BuildFromRequest(b io.ReadCloser, p *project.Project, c *Commit) (*Build, error) {
// 	reqBuild := RequestBuild{}

// 	err := json.NewDecoder(b).Decode(&reqBuild)
// 	if err != nil {
// 		return nil, err
// 	}

// 	reqBuild.TaskName = strings.TrimSpace(reqBuild.TaskName)
// 	for i, rP := range reqBuild.Parameters {
// 		reqBuild.Parameters[i].Value = strings.TrimSpace(rP.Value)
// 	}

// 	// err = r.buildValidator.Validate(&reqBuild)

// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	task, err := r.CommitManager.GetTaskForCommitInProject(c, p, reqBuild.TaskName)
// 	if err != nil {
// 		log.Fatal(err)
// 		return nil, err
// 	}

// 	setTaskParametersFromRequest(task, reqBuild.Parameters)

// 	build := NewBuild(p.ID, c.Hash, task)

// 	build.ID = r.CommitManager.GetNextBuildID(p, c)

// 	return &build, nil
// }

// func setTaskParametersFromRequest(t *velocity.Task, reqParams []RequestParameter) {
// 	for _, reqParam := range reqParams {
// 		if param, ok := t.Parameters[reqParam.Name]; ok {
// 			param.Value = reqParam.Value
// 			t.Parameters[reqParam.Name] = param
// 		}
// 	}
// }
