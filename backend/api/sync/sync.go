package sync

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/velocity-ci/velocity/backend/api/domain/branch"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
)

func sync(
	p *project.Project,
	projectManager project.Repository,
	commitManager commit.Repository,
	branchManager branch.Repository,
	taskManager task.Repository,
) {
	repo, dir, err := GitClone(p, false, false, true, velocity.NewBlankWriter())
	if err != nil {
		log.Fatal(err)
		return
	}
	defer os.RemoveAll(dir) // clean up

	refIter, err := repo.References()
	if err != nil {
		log.Fatal(err)
		return
	}
	w, err := repo.Worktree()
	if err != nil {
		log.Fatal(err)
		return
	}
	for {
		r, err := refIter.Next()
		if err != nil {
			break
		}

		fmt.Println(r)
		gitCommit, err := repo.CommitObject(r.Hash())

		if err != nil {
			break
		}

		mParts := strings.Split(gitCommit.Message, "-----END PGP SIGNATURE-----")
		message := mParts[0]
		if len(mParts) > 1 {
			message = mParts[1]
		}

		branchName := strings.Join(strings.Split(r.Name().Short(), "/")[1:], "/")
		if branchName != "" {
			b := branch.NewBranch(p, branchName)
			branchManager.Save(b)

			c := commit.NewCommit(
				p,
				gitCommit.Hash.String(),
				strings.TrimSpace(message),
				gitCommit.Author.Email,
				gitCommit.Committer.When,
				*b,
			)

			commitManager.Save(c)

			err = w.Checkout(&git.CheckoutOptions{
				Hash: gitCommit.Hash,
			})

			if err != nil {
				fmt.Println(err)
			}

			SHA := r.Hash().String()
			shortSHA := SHA[:7]
			describe := shortSHA

			gitParams := map[string]velocity.Parameter{
				"GIT_SHA": velocity.Parameter{
					Value: SHA,
				},
				"GIT_SHORT_SHA": velocity.Parameter{
					Value: shortSHA,
				},
				"GIT_BRANCH": velocity.Parameter{
					Value: branchName,
				},
				"GIT_DESCRIBE": velocity.Parameter{
					Value: describe,
				},
			}

			if _, err := os.Stat(fmt.Sprintf("%s/tasks/", dir)); err == nil {
				filepath.Walk(fmt.Sprintf("%s/tasks/", dir), func(path string, f os.FileInfo, err error) error {
					if !f.IsDir() && strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml") {
						taskYml, _ := ioutil.ReadFile(fmt.Sprintf("%s/tasks/%s", dir, f.Name()))
						t := velocity.ResolveTaskFromYAML(string(taskYml), gitParams)
						apiTask := task.NewTask(p, c, t)
						taskManager.Save(apiTask)
					}
					return nil
				})
			}
		}

	}

	p.UpdatedAt = time.Now()
	p.Synchronising = false
	projectManager.Save(p)
}
