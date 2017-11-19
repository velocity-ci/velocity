package main

import "time"

type GORMBuild struct {
	ID         string
	ProjectID  string
	CommitID   string
	TaskID     string
	Status     string // waiting, running, failed, success
	Parameters []Parameter
}

type Parameter struct {
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
