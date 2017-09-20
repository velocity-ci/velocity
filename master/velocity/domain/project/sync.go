package project

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/domain/task"
	git "gopkg.in/src-d/go-git.v4"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

func sync(p *domain.Project, m *BoltManager) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("velocity_%s", p.ID))
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up

	// Clones the repository into the given dir, just as a normal git clone does
	signer, _ := ssh.ParsePrivateKey([]byte(p.PrivateKey))
	auth := &gitssh.PublicKeys{User: "git", Signer: signer}
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:   p.Repository,
		Depth: 1,
		Auth:  auth,
	})

	if err != nil {
		log.Fatal(err)
	}

	refIter, err := repo.References()
	w, err := repo.Worktree()
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

		c := domain.Commit{
			Hash:    commit.Hash.String(),
			Message: strings.TrimSpace(message),
			Author:  commit.Author.Email,
			Date:    commit.Committer.When,
		}

		m.SaveCommitForProject(p, &c)

		err = w.Checkout(&git.CheckoutOptions{
			Hash: commit.Hash,
		})

		if err != nil {
			fmt.Println(err)
		}

		SHA := r.Hash().String()
		shortSHA := SHA[:7]
		branch := r.Name().Short()
		describe := shortSHA

		gitParams := []task.Parameter{
			task.Parameter{
				Name:  "GIT_SHA",
				Value: SHA,
			},
			task.Parameter{
				Name:  "GIT_SHORT_SHA",
				Value: shortSHA,
			},
			task.Parameter{
				Name:  "GIT_BRANCH",
				Value: branch,
			},
			task.Parameter{
				Name:  "GIT_DESCRIBE",
				Value: describe,
			},
		}

		filepath.Walk(fmt.Sprintf("%s/tasks/", dir), func(path string, f os.FileInfo, err error) error {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml") {
				taskYml, _ := ioutil.ReadFile(fmt.Sprintf("%s/tasks/%s", dir, f.Name()))
				task := task.ResolveTaskFromYAML(string(taskYml), gitParams)
				m.SaveTaskForCommitInProject(&task, &c, p)
			}
			return nil
		})

	}

	p.UpdatedAt = time.Now()
	p.Synchronising = false
	m.Save(p)

}
