package git

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/velocity-ci/velocity/master/velocity/project"
	"github.com/velocity-ci/velocity/master/velocity/project/git/repository"
	git "gopkg.in/src-d/go-git.v4"
)

type Project struct {
	project.Project
	Repository repository.Repository
}

func (p *Project) Clone(bare bool) (*git.Repository, string, error) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("velocity_%s", project.IdFromName(p.Name)))
	if err != nil {
		log.Fatal(err)
		return nil, "", err
	}

	authMethod, err := p.Repository.GetAuthMethod()
	if err != nil {
		return nil, "", err
	}

	repo, err := git.PlainClone(dir, bare, &git.CloneOptions{
		URL:   p.Repository.GetAddress(),
		Depth: 1,
		Auth:  authMethod,
	})

	if err != nil {
		os.RemoveAll(dir)
		return nil, "", err
	}

	return repo, dir, nil
}
