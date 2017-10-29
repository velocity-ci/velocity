package project

import (
	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
)

type SyncManager struct {
	Sync func(p *velocity.Project, bare bool, full bool, submodule bool, emitter velocity.Emitter) (*git.Repository, string, error)
}

func NewSyncManager(cloneFunc func(p *velocity.Project, bare bool, full bool, submodule bool, emitter velocity.Emitter) (*git.Repository, string, error)) *SyncManager {
	return &SyncManager{
		Sync: cloneFunc,
	}
}
