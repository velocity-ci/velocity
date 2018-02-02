package persistence

import (
	"github.com/velocity-ci/velocity/backend/architect/domain"
	"github.com/velocity-ci/velocity/backend/architect/domain/persistence/db"
)

func SaveProject(p *domain.Project) error {
	return db.SaveProject(p)
}
