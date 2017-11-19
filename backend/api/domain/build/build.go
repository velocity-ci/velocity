package build

import (
	"time"

	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/project"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type Repository interface {
	Save(b *Build) *Build
	Delete(b *Build)
	GetByProjectAndCommitAndID(p *project.Project, c *commit.Commit, id string) (*Build, error)
	GetAllByProject(p *project.Project, q Query) ([]*Build, uint64)
	GetAllByProjectAndCommit(p *project.Project, c *commit.Commit) ([]*Build, uint64)
}

type Query struct {
	Amount uint64
	Page   uint64
}

type Build struct {
	ID         string
	Status     string
	Parameters []velocity.Parameter
	BuildSteps []BuildStep
}

// /v1/projects/velocity/commits/abcdef/builds/3/steps/2/containerTwo/logs OutputStream/Lines

type BuildStep struct {
	ID            string
	Status        string
	OutputStreams []string
}

type StreamLine struct {
	OutputStreamID string
	LineNumber     uint64
	Timestamp      time.Time
	Output         string
}
