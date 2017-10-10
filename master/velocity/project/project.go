package project

import (
	"strings"
	"time"
)

type Project struct {
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type requestProject struct {
	Name       string `json:"name" validate:"required,min=3,max=128,projectUnique"`
	Repository string `json:"repository" validate:"required,min=8,max=128"`
	PrivateKey string `json:"key"`
}

func IdFromName(name string) string {
	return strings.Replace(strings.ToLower(name), " ", "-", -1)
}
