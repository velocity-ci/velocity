package project

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/velocity-ci/velocity/master/velocity/domain"
	git "gopkg.in/src-d/go-git.v4"
)

func sync(p *domain.Project) {
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

		c := domain.Commit{
			Hash:    commit.Hash.String(),
			Author:  commit.Author.Email,
			Message: commit.Message,
			Date:    commit.Committer.When,
		}

		p.Commits = append(p.Commits, c)

		fmt.Println(c)

	}

	// Prints the content of the CHANGELOG file from the cloned repository
	// changelog, err := os.Open(filepath.Join(dir, "README.md"))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// io.Copy(os.Stdout, changelog)
}
