package commit

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/velocity-ci/velocity/backend/api/project"
	"github.com/velocity-ci/velocity/backend/task"
	git "gopkg.in/src-d/go-git.v4"
)

func sync(p *project.Project, projectManager *project.Manager, commitManager *Manager) {
	repo, dir, err := project.Clone(*p, false)
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
		commit, err := repo.CommitObject(r.Hash())

		if err != nil {
			break
		}

		mParts := strings.Split(commit.Message, "-----END PGP SIGNATURE-----")
		message := mParts[0]
		if len(mParts) > 1 {
			message = mParts[1]
		}

		branch := strings.Join(strings.Split(r.Name().Short(), "/")[1:], "/")
		err = commitManager.SaveBranchForProject(p, branch)
		if err != nil {
			log.Fatal(err)
		}

		c := Commit{
			Branch:  branch,
			Hash:    commit.Hash.String(),
			Message: strings.TrimSpace(message),
			Author:  commit.Author.Email,
			Date:    commit.Committer.When,
		}

		err = commitManager.SaveCommitForProject(p, &c)
		if err != nil {
			log.Fatal(err)
		}

		err = w.Checkout(&git.CheckoutOptions{
			Hash: commit.Hash,
		})

		if err != nil {
			fmt.Println(err)
		}

		SHA := r.Hash().String()
		shortSHA := SHA[:7]
		describe := shortSHA

		gitParams := map[string]task.Parameter{
			"GIT_SHA": task.Parameter{
				Value: SHA,
			},
			"GIT_SHORT_SHA": task.Parameter{
				Value: shortSHA,
			},
			"GIT_BRANCH": task.Parameter{
				Value: branch,
			},
			"GIT_DESCRIBE": task.Parameter{
				Value: describe,
			},
		}

		if _, err := os.Stat(fmt.Sprintf("%s/tasks/", dir)); err == nil {
			filepath.Walk(fmt.Sprintf("%s/tasks/", dir), func(path string, f os.FileInfo, err error) error {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml") {
					taskYml, _ := ioutil.ReadFile(fmt.Sprintf("%s/tasks/%s", dir, f.Name()))
					t := task.ResolveTaskFromYAML(string(taskYml), gitParams)
					err = commitManager.SaveTaskForCommitInProject(&t, &c, p)
					if err != nil {
						log.Fatal(err)
					}
				}
				return nil
			})
		}
	}

	p.UpdatedAt = time.Now()
	p.Synchronising = false
	projectManager.Save(p)

}
