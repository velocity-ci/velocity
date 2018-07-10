package sync

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
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
		velocity.GetLogger().Info("found commit",
			zap.String("project", p.Slug),
			zap.String("branch", branchName),
			zap.String("message", rawCommit.Message),
			zap.Time("at", rawCommit.AuthorDate),
		)
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
			velocity.GetLogger().Info("created commit",
				zap.String("project", p.Slug),
				zap.String("sha", c.Hash),
				zap.String("branch", branchName),
			)

			err = repo.Checkout(rawCommit.SHA)

			if err != nil {
				velocity.GetLogger().Error("error", zap.Error(err))
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
							velocity.GetLogger().Error("error", zap.Error(err))
						} else {
							taskManager.Create(c, &t, velocity.NewSetup())
							velocity.GetLogger().Info("created task",
								zap.String("project", p.Slug),
								zap.String("sha", c.Hash),
								zap.String("task", t.Name),
							)
						}
					}
					return nil
				})
			}
		} else if !branchManager.HasCommit(b, c) {
			velocity.GetLogger().Info("added commit to branch",
				zap.String("project", p.Slug),
				zap.String("sha", c.Hash),
				zap.String("branch", b.Name),
			)
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
