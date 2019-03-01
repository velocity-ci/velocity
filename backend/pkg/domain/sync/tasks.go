package sync

import (
	"fmt"
	"os"

	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
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

		rawCommit, err := repo.GetCommitAtHeadOfBranch(branchName)
		if err != nil {
			continue
		}
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
				tasks, err := velocity.GetTasksFromCurrentDir()
				if err != nil {
					return err
				}
				for _, task := range tasks {
					if len(task.ValidationErrors) > 0 {
						velocity.GetLogger().Warn(
							"skipping task because of validation errors",
							zap.String("task", task.Name),
							zap.Strings("validationErrors", task.ValidationErrors),
						)
						continue
					}
					taskManager.Create(c, task, velocity.NewSetup())
					velocity.GetLogger().Info("created task",
						zap.String("project", p.Slug),
						zap.String("sha", c.Hash),
						zap.String("task", task.Name),
					)

				}
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
