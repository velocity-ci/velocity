package sync

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/glog"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	yaml "gopkg.in/yaml.v2"
)

func syncRepository(p *project.Project, repo *velocity.RawRepository) (*project.Project, error) {

	err := repo.Checkout(repo.GetDefaultBranch())
	if err != nil {
		return p, err
	}

	repoConfigPath := fmt.Sprintf("%s/.velocity.yaml", repo.Directory)
	if _, err := os.Stat(repoConfigPath); err != nil {
		glog.Infof("No repository config found (.velocity.yaml)")
		return p, nil
	}

	repoYaml, _ := ioutil.ReadFile(repoConfigPath)

	err = yaml.Unmarshal(repoYaml, &p.RepositoryConfig)
	if err != nil {
		return p, err
	}

	return p, nil
}
