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
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

func clone(name string, repositoryAddress string, key string, bare bool) (*git.Repository, string, error) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("velocity_%s", idFromName(name)))
	if err != nil {
		log.Fatal(err)
		return nil, "", err
	}

	isGit := repositoryAddress[:3] == "git"

	var auth transport.AuthMethod

	if isGit {
		log.Printf("git repository: %s", repositoryAddress)
		signer, err := ssh.ParsePrivateKey([]byte(key))
		if err != nil {
			os.RemoveAll(dir)
			return nil, "", SSHKeyError(err.Error())
		}
		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}

	repo, err := git.PlainClone(dir, bare, &git.CloneOptions{
		URL:   repositoryAddress,
		Depth: 1,
		Auth:  auth,
	})

	if err != nil {
		os.RemoveAll(dir)
		return nil, "", err
	}

	return repo, dir, nil
}

type SSHKeyError string

func (s SSHKeyError) Error() string {
	return string(s)
}

func sync(p *domain.Project, m *BoltManager) {
	repo, dir, err := clone(p.Name, p.Repository, p.PrivateKey, false)
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

		c := domain.Commit{
			Branch:  branch,
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

		if _, err := os.Stat(fmt.Sprintf("%s/tasks/", dir)); err == nil {
			filepath.Walk(fmt.Sprintf("%s/tasks/", dir), func(path string, f os.FileInfo, err error) error {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml") {
					taskYml, _ := ioutil.ReadFile(fmt.Sprintf("%s/tasks/%s", dir, f.Name()))
					t := task.ResolveTaskFromYAML(string(taskYml), gitParams)
					m.SaveTaskForCommitInProject(&t, &c, p)
				}
				return nil
			})
		}
	}

	p.UpdatedAt = time.Now()
	p.Synchronising = false
	m.Save(p)

}
