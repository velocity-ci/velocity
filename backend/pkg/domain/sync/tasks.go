package sync

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	yaml "gopkg.in/yaml.v2"
)

func syncTasks(
	p *project.Project,
	repo *velocity.RawRepository,
	taskManager *task.Manager,
	branchManager *githistory.BranchManager,
	commitManager *githistory.CommitManager,
) error {
	branches := repo.GetBranches()

	for _, branchName := range branches {
		b, err := branchManager.GetByProjectAndName(p, branchName)
		if err != nil {
			b = branchManager.Create(p, branchName)
		}

		rawCommit := repo.GetCommitAtHeadOfBranch(branchName)
		glog.Infof("got commit for %s:%s - %s @ %s", branchName, rawCommit.SHA, rawCommit.Message, rawCommit.AuthorDate)
		c, err := commitManager.GetByProjectAndHash(p, rawCommit.SHA)

		if err != nil {
			c = commitManager.Create(
				b,
				p,
				rawCommit.SHA,
				rawCommit.Message,
				rawCommit.AuthorEmail,
				rawCommit.AuthorDate,
				rawCommit.Signed,
			)
			glog.Infof("\tcreated commit %s on %s", c.Hash, b.Name)

			err = repo.Checkout(rawCommit.SHA)

			if err != nil {
				glog.Error(err)
				break
			}

			if _, err := os.Stat(fmt.Sprintf("%s/tasks/", repo.Directory)); err == nil {
				os.Chdir(repo.Directory)
				filepath.Walk(fmt.Sprintf("%s/tasks/", repo.Directory), func(path string, f os.FileInfo, err error) error {
					if !f.IsDir() && strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml") {
						taskYml, _ := ioutil.ReadFile(fmt.Sprintf("%s/tasks/%s", repo.Directory, f.Name()))
						var t velocity.Task
						err := yaml.Unmarshal(taskYml, &t)
						if err != nil {
							glog.Error(err)
						} else {
							taskManager.Create(c, &t, velocity.NewSetup())
							glog.Infof("\tcreated task %s for %s", t.Name, c.Hash)
						}
					}
					return nil
				})
			}
		} else if !branchManager.HasCommit(b, c) {
			glog.Infof("\tadded commit %s to %s", c.Hash, b.Name)
			commitManager.AddCommitToBranch(c, b)
		}
	}

	// Set remaining local branches as inactive.
	allKnownBranches, _ := branchManager.GetAllForProject(p, &domain.PagingQuery{Limit: 100, Page: 1})

	localOnlyBranches := removeRemoteBranches(allKnownBranches, branches)
	for _, b := range localOnlyBranches {
		b.Active = false
		branchManager.Update(b)
	}

	return nil
}

func removeRemoteBranches(haystack []*githistory.Branch, names []string) (r []*githistory.Branch) {
	for _, b := range haystack {
		found := false
		for _, n := range names {
			if b.Name == n {
				found = true
				break
			}
		}
		if !found {
			r = append(r, b)
		}
	}
	return r
}
