package db

import (
	"github.com/docker/go/canonical/json"
	"github.com/velocity-ci/velocity/backend/architect/domain"
)

type project struct {
	UUID           string `gorm:"primary_key"`
	Name           string `gorm:"not null"`
	RepoConfigJSON []byte
}

func fromDomainProject(p *domain.Project) project {
	repoConfigJSON, _ := json.Marshal(p.Config)
	return project{
		UUID:           p.UUID,
		Name:           p.Name,
		RepoConfigJSON: repoConfigJSON,
	}
}

func SaveProject(p *domain.Project) error {
	tx := db.Begin()

	gP := fromDomainProject(p)

	tx.
		Where(project{UUID: gP.UUID}).
		Assign(&gP).
		FirstOrCreate(&gP)

	return tx.Commit().Error
}
