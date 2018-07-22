package sync

import (
	"fmt"
	"io/ioutil"
	"os"

	"go.uber.org/zap"

	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	yaml "gopkg.in/yaml.v2"
)

func syncRepository(p *project.Project, repo *velocity.RawRepository) (*project.Project, error) {
	err := repo.Checkout(repo.GetDefaultBranch())
	if err != nil {
		return p, err
	}

	repoConfigPathYaml := fmt.Sprintf("%s/.velocity.yaml", repo.Directory)
	repoConfigPathYml := fmt.Sprintf("%s/.velocity.yml", repo.Directory)
	repoConfig := ""
	if _, err := os.Stat(repoConfigPathYaml); err == nil {
		repoConfig = repoConfigPathYaml
	} else if _, err := os.Stat(repoConfigPathYml); err == nil {
		repoConfig = repoConfigPathYml
	} else {
		velocity.GetLogger().Warn("could not find repository config .velocity.yaml or .velocity.yml", zap.Error(err))
		return p, nil
	}

	repoYaml, _ := ioutil.ReadFile(repoConfig)

	err = yaml.Unmarshal(repoYaml, &p.RepositoryConfig)
	if err != nil {
		return p, err
	}

	return p, nil
}
