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

	// err := repo.Checkout(repo.GetDefaultBranch())
	err := repo.Checkout("feature/config-yaml")
	if err != nil {
		return p, err
	}

	repoConfigPath := fmt.Sprintf("%s/.velocity.yaml", repo.Directory)
	if _, err := os.Stat(repoConfigPath); err != nil {
		velocity.GetLogger().Debug("could not find repository config .velocity.yaml", zap.Error(err))
		return p, nil
	}

	repoYaml, _ := ioutil.ReadFile(repoConfigPath)

	err = yaml.Unmarshal(repoYaml, &p.RepositoryConfig)
	if err != nil {
		return p, err
	}

	return p, nil
}
