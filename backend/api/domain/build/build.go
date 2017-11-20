package build

import (
	"time"

	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
	"github.com/velocity-ci/velocity/backend/project"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type Repository interface {
	SaveToProjectAndCommit(p *project.Project, c *commit.Commit, b *Build) *Build
	DeleteFromProjectAndCommit(p *project.Project, c *commit.Commit, b *Build)
	GetByProjectAndCommitAndID(p *project.Project, c *commit.Commit, id string) (*Build, error)
	// Order timestamp descending
	GetAllByProject(p *project.Project, q Query) ([]*Build, uint64)
	// Order timestamp descending
	GetAllByProjectAndCommit(p *project.Project, c *commit.Commit) ([]*Build, uint64)
}

type Query struct {
	Amount uint64
	Page   uint64
}

type Build struct {
	ID         string
	Project    project.Project
	Commit     commit.Commit
	Task       task.Task
	Status     string
	Parameters []velocity.Parameter
}

// /v1/projects/velocity/commits/abcdef/builds/3/steps/2/containerTwo/logs OutputStream/Lines

type BuildStep struct {
	ID            string
	Status        string
	OutputStreams []OutputStream
}

type OutputStream struct {
	Name        string
	BuildStepID string
	Lines       []StreamLine
}

type StreamLine struct {
	OutputStreamID string
	LineNumber     uint64
	Timestamp      time.Time
	Output         string
}
