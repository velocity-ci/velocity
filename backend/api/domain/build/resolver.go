package build

import (
	"io"
	"log"
	"strings"

	"github.com/docker/go/canonical/json"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type Resolver struct {
	taskManager task.Repository
	// BuildValidator *BuildValidator
}

func NewResolver(taskManager *task.Manager) *Resolver {
	return &Resolver{
		taskManager: taskManager,
		// BuildValidator: buildValidator,
	}
}

func (r *Resolver) BuildFromRequest(b io.ReadCloser, p *project.Project, c *commit.Commit) (*Build, error) {
	reqBuild := RequestBuild{}

	err := json.NewDecoder(b).Decode(&reqBuild)
	if err != nil {
		return nil, err
	}

	for i, rP := range reqBuild.Parameters {
		reqBuild.Parameters[i].Value = strings.TrimSpace(rP.Value)
	}

	// err = r.buildValidator.Validate(&reqBuild)

	// if err != nil {
	// 	return nil, err
	// }

	task, err := r.taskManager.GetByProjectAndCommitAndID(p, c, reqBuild.TaskID)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	setTaskParametersFromRequest(&task.VTask, reqBuild.Parameters)

	build := NewBuild(p, c, task, task.VTask.Parameters)

	return build, nil
}

func setTaskParametersFromRequest(t *velocity.Task, reqParams []RequestParameter) {
	for _, reqParam := range reqParams {
		if param, ok := t.Parameters[reqParam.Name]; ok {
			param.Value = reqParam.Value
			t.Parameters[reqParam.Name] = param
		}
	}
}
