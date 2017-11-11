package main

import "time"

type GORMProject struct {
	ID string `gorm:"primary_key"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Synchronising bool
}

// GORMBranch - Has a many-to-many relationship with GORMCommits
type GORMBranch struct {
	ID        string
	ProjectID string
}

type GORMBranchCommit struct {
	BranchID string
	CommitID string
}

// GORMCommit - Has a many-to-many relationship with GORMBranch
type GORMCommit struct {
	ID        string // Hash
	ProjectID string
}

type GORMTask struct {
	TaskID   string
	CommitID string // Hash
	Name     string
}

type GORMStep struct {
	TaskID string
}

type GORMBuild struct {
	ID        string
	ProjectID string
	CommitID  string
	TaskID    string
	Status    string // waiting, running, failed, success
}

type GORMBuildStep struct {
	ID      string
	BuildID string
	Status  string // running, failed, success
}

type GORMOutputStream struct {
	ID          string
	Name        string
	BuildStepID string
	OutputFile  string
}

type StreamLine struct {
	OutputStreamID string
	LineNumber     uint64
	Timestamp      time.Time
	Output         string
}
