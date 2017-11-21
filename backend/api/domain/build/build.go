package build

import (
	"time"

	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
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

	// BuildSteps
	SaveBuildStep(bS *BuildStep) *BuildStep
	GetBuildStepsForBuild(b *Build) ([]*BuildStep, uint64)

	// OutputStreams
	SaveOutputStream(oS *OutputStream) *OutputStream
	GetOutputStreamsForBuildStep(bS *BuildStep) ([]*OutputStream, uint64)
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

// /v1/projects/velocity/commits/abcdef/builds/<3>/steps/<2>/streams/<containerTwo>/logs OutputStream/Lines

type BuildStep struct {
	ID     string
	Build  Build
	Status string
}

type OutputStream struct {
	ID        string
	BuildStep BuildStep
	Name      string
	Path      string
}

type StreamLine struct {
	OutputStreamID string
	LineNumber     uint64
	Timestamp      time.Time
	Output         string
}
