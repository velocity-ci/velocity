package persistence

import (
	"github.com/velocity-ci/velocity/backend/architect/domain"
	"github.com/velocity-ci/velocity/backend/architect/domain/persistence/db"
)

func SaveKnownHost(k *domain.KnownHost) error {
	if err := db.SaveKnownHost(k); err != nil {
		return err
	}
	// update file
	return nil
}

func GetKnownHosts() ([]*domain.KnownHost, int) {
	return db.GetKnownHosts()
}
