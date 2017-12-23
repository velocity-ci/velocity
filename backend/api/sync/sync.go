package sync

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
	"github.com/velocity-ci/velocity/backend/api/websocket"
	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
)

func finishProjectSync(p project.Project, projectManager project.Repository) {
	p.UpdatedAt = time.Now()
	p.Synchronising = false
	projectManager.Update(p)
}

func sync(
	p project.Project,
	projectManager project.Repository,
	commitManager commit.Repository,
	taskManager task.Repository,
	websocketManager *websocket.Manager,
) {
	defer finishProjectSync(p, projectManager)
	repo, dir, err := velocity.GitClone(&p.Repository, false, false, true, velocity.NewBlankWriter())
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
	remoteBranchNames := []string{}
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
			remoteBranchNames = append(remoteBranchNames, branchName)
			b, err := commitManager.GetBranchByProjectIDAndName(p.ID, branchName)
			if err != nil {
				b = commit.NewBranch(p.ID, branchName)
				commitManager.CreateBranch(b)
			}

			c, err := commitManager.GetCommitByProjectIDAndCommitHash(p.ID, gitCommit.Hash.String())
			if err != nil {
				c = commit.NewCommit(
					p.ID,
					gitCommit.Hash.String(),
					strings.TrimSpace(message),
					gitCommit.Author.Email,
					gitCommit.Committer.When,
					[]commit.Branch{b},
				)
				commitManager.CreateCommit(c)

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
							apiTask := task.NewTask(c.ID, t)
							taskManager.Create(apiTask)
						}
						return nil
					})
				}
			}
		}
	}

	// Set remaining local branches as inactive.
	allKnownBranches, _ := commitManager.GetAllBranchesByProjectID(p.ID, commit.BranchQuery{})
	localOnlyBranches := removeRemoteBranches(allKnownBranches, remoteBranchNames)
	for _, b := range localOnlyBranches {
		b.Active = false
		commitManager.UpdateBranch(b)
	}

}

func removeRemoteBranches(haystack []commit.Branch, names []string) []commit.Branch {
	returnBranches := []commit.Branch{}
	for _, b := range haystack {
		found := false
		for _, n := range names {
			if b.Name == n {
				found = true
				break
			}
		}
		if !found {
			returnBranches = append(returnBranches, b)
		}
	}
	return returnBranches
}
