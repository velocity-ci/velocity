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
	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
	yaml "gopkg.in/yaml.v2"
)

func finishProjectSync(p *project.Project, projectManager *project.Manager) {
	p.UpdatedAt = time.Now()
	p.Synchronising = false
	projectManager.Update(p)
}

func sync(
	p *project.Project,
	taskManager *Manager,
	// websocketManager *websocket.Manager,
) {
	defer finishProjectSync(p, taskManager.projectManager)
	repo, dir, err := velocity.GitClone(&p.Config, false, false, true, velocity.NewBlankEmitter().GetStreamWriter("clone"))
	if err != nil {
		logrus.Error(err)
		return
	}
	defer os.RemoveAll(dir) // clean up

	refIter, err := repo.References()
	if err != nil {
		logrus.Error(err)
		return
	}
	w, err := repo.Worktree()
	if err != nil {
		logrus.Error(err)
		return
	}
	remoteBranchNames := []string{}
	for {
		r, err := refIter.Next()
		if err != nil {
			logrus.Error(err)
			break
		}

		fmt.Println(r)
		gitCommit, err := repo.CommitObject(r.Hash())

		if err != nil {
			logrus.Error(err)
			break
		}

		mParts := strings.Split(gitCommit.Message, "-----END PGP SIGNATURE-----")
		message := mParts[0]
		if len(mParts) > 1 {
			message = mParts[1]
		}

		branchName := strings.Join(strings.Split(r.Name().Short(), "/")[1:], "/")
		if branchName != "" {
			remoteBranchNames = append(remoteBranchNames, branchName)
			b, err := taskManager.branchManager.GetByProjectAndName(p, branchName)
			if err != nil {
				b = taskManager.branchManager.Create(p, branchName)
			}

			c, err := taskManager.commitManager.GetByProjectAndHash(p, gitCommit.Hash.String())

			if err != nil {
				c = taskManager.commitManager.Create(
					b,
					p,
					gitCommit.Hash.String(),
					strings.TrimSpace(message),
					gitCommit.Author.Email,
					gitCommit.Committer.When,
				)
			}

			if !taskManager.branchManager.HasCommit(b, c) {
				err = w.Checkout(&git.CheckoutOptions{
					Hash: gitCommit.Hash,
				})

				if err != nil {
					logrus.Error(err)
					break
				}

				xd, _ := os.Getwd()
				if _, err := os.Stat(fmt.Sprintf("%s/tasks/", dir)); err == nil {
					os.Chdir(dir)
					filepath.Walk(fmt.Sprintf("%s/tasks/", dir), func(path string, f os.FileInfo, err error) error {
						if !f.IsDir() && strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml") {
							taskYml, _ := ioutil.ReadFile(fmt.Sprintf("%s/tasks/%s", dir, f.Name()))
							var t velocity.Task
							err := yaml.Unmarshal(taskYml, &t)
							if err != nil {
								logrus.Error(err)
							} else {
								taskManager.Create(c, &t, velocity.NewSetup())
							}
						}
						return nil
					})
					os.Chdir(xd)
				}
			}
		}
	}

	// Set remaining local branches as inactive.
	allKnownBranches, _ := taskManager.branchManager.GetAllForProject(p, &domain.PagingQuery{Limit: 100, Page: 1})

	localOnlyBranches := removeRemoteBranches(allKnownBranches, remoteBranchNames)
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
