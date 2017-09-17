package project

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/velocity-ci/velocity/master/velocity/domain"
	git "gopkg.in/src-d/go-git.v4"
)

func sync(p *domain.Project, m *Manager) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("velocity_%s", p.ID))
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up

	// Clones the repository into the given dir, just as a normal git clone does
	repo, err := git.PlainClone(dir, true, &git.CloneOptions{
		URL: p.Repository,
	})

	if err != nil {
		log.Fatal(err)
	}

	refIter, err := repo.References()
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
	}

	p.Synchronising = false
	m.Save(p)

}
