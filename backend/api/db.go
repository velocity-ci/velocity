package main

import "time"

type GORMProject struct {
	ID        string    `gorm:"primary_key"`
	CreatedAt time.Time `json:"createdAt"`
}

type GORMCommit struct {
}

type GORMTask struct {
}

type GORMStep struct {
}

type GORMBuild struct {
}
