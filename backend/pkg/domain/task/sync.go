package task

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain"

	"github.com/Sirupsen/logrus"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	yaml "gopkg.in/yaml.v2"
)

func finishProjectSync(p *project.Project, projectManager *project.Manager) {
	p.UpdatedAt = time.Now()
	p.Synchronising = false
	projectManager.Update(p)
	logrus.Infof("finished synchronising project %s", p.Slug)
}

func sync(
	p *project.Project,
	taskManager *Manager,
) {
	logrus.Infof("synchronising project %s", p.Slug)
	xd, _ := os.Getwd()
	defer os.Chdir(xd)
	defer finishProjectSync(p, taskManager.projectManager)
	repo, err := velocity.Clone(&p.Config, false, true, false, velocity.NewBlankEmitter().GetStreamWriter("clone"))
	if err != nil {
		logrus.Errorf("could not clone repository %s", err)
		return
	}
	defer os.RemoveAll(repo.Directory) // clean up

	branches := repo.GetBranches()

	for _, branchName := range branches {
		b, err := taskManager.branchManager.GetByProjectAndName(p, branchName)
		if err != nil {
			b = taskManager.branchManager.Create(p, branchName)
		}

		rawCommit := repo.GetCommitAtHeadOfBranch(branchName)
		logrus.Infof("got commit for %s:%s - %s @ %s", branchName, rawCommit.SHA, rawCommit.Message, rawCommit.AuthorDate)
		c, err := taskManager.commitManager.GetByProjectAndHash(p, rawCommit.SHA)

		if err != nil {
			c = taskManager.commitManager.Create(
				b,
				p,
				rawCommit.SHA,
				rawCommit.Message,
				rawCommit.AuthorEmail,
				rawCommit.AuthorDate,
			)
			logrus.Infof("\tcreated commit %s on %s", c.Hash, b.Name)

			err = repo.Checkout(rawCommit.SHA)

			if err != nil {
				logrus.Error(err)
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
							logrus.Error(err)
						} else {
							taskManager.Create(c, &t, velocity.NewSetup())
							logrus.Infof("\tcreated task %s for %s", t.Name, c.Hash)
						}
					}
					return nil
				})
			}
		} else if !taskManager.branchManager.HasCommit(b, c) {
			logrus.Infof("\tadded commit %s to %s", c.Hash, b.Name)
			taskManager.commitManager.AddCommitToBranch(c, b)
		}
	}

	// Set remaining local branches as inactive.
	allKnownBranches, _ := taskManager.branchManager.GetAllForProject(p, &domain.PagingQuery{Limit: 100, Page: 1})

	localOnlyBranches := removeRemoteBranches(allKnownBranches, branches)
	for _, b := range localOnlyBranches {
		b.Active = false
		taskManager.branchManager.Update(b)
	}

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
