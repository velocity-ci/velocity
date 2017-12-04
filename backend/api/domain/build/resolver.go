package build

import (
	"io"
	"log"
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

	// err = r.buildValidator.Validate(&reqBuild)

	// if err != nil {
	// 	return nil, err
	// }

	setTaskParametersFromRequest(&t, reqBuild.Parameters)

	cm, err := r.commitManager.GetCommitByCommitID(t.CommitID)
	if err != nil {
		log.Printf("could not find commit %s?!?!", t.CommitID)
	}

	build := NewBuild(cm.ProjectID, t.ID, t.Parameters)

	return build, nil
}

func setTaskParametersFromRequest(t *task.Task, reqParams []RequestParameter) {
	for _, reqParam := range reqParams {
		if param, ok := t.Parameters[reqParam.Name]; ok {
			param.Value = reqParam.Value
			t.Parameters[reqParam.Name] = param
		}
	}
}
