package sync

import (
	"fmt"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type Manager struct {
	projectManager *project.Manager
	taskManager    *task.Manager
	branchManager  *githistory.BranchManager
	commitManager  *githistory.CommitManager
}

func NewManager(
	projectManager *project.Manager,
	taskManager *task.Manager,
	branchManager *githistory.BranchManager,
	commitManager *githistory.CommitManager,
) *Manager {
	return &Manager{
		projectManager: projectManager,
		taskManager:    taskManager,
		branchManager:  branchManager,
		commitManager:  commitManager,
	}
}

func (m *Manager) Sync(p *project.Project) (*project.Project, error) {
	if p.Synchronising {
		return nil, fmt.Errorf("already syncronising")
	}

	p.Synchronising = true
	if err := m.projectManager.Update(p); err != nil {
		return nil, err
	}

	go sync(p, m)

	return p, nil
}

func sync(p *project.Project, m *Manager) {
	glog.Infof("synchronising project %s", p.Slug)
	xd, _ := os.Getwd()
	defer os.Chdir(xd)
	defer finishSync(p, m)
	// clone
	repo, err := velocity.Clone(&p.Config, velocity.NewBlankEmitter().GetStreamWriter("clone"), &velocity.CloneOptions{
		Bare:      false,
		Full:      false,
		Submodule: true,
	})
	if err != nil {
		glog.Errorf("could not clone repository %s", err)
		return
	}
	defer os.RemoveAll(repo.Directory) // clean up
	// sync repository
	p, err = syncRepository(p, repo)

	// clone further
	if p.RepositoryConfig.Git.Depth > 1 {

	}

	// sync tasks
	err = syncTasks(p, repo, m.taskManager, m.branchManager, m.commitManager)
}

func finishSync(p *project.Project, m *Manager) {
	p.UpdatedAt = time.Now()
	p.Synchronising = false
	m.projectManager.Update(p)
	glog.Infof("finished synchronising project %s", p.Slug)
}
