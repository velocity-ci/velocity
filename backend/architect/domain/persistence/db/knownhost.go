package db

import (
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/architect/domain"
	"golang.org/x/crypto/ssh"
)

type knownhost struct {
	UUID  string `gorm:"primary_key"`
	Entry string `gorm:"not null"`
}

func (g knownhost) toDomainKnownhost() *domain.KnownHost {
	k := &domain.KnownHost{
		Entry: g.Entry,
	}

	_, hosts, pubKey, comment, _, _ := ssh.ParseKnownHosts([]byte(g.Entry))

	k.UUID = uuid.NewV1().String()
	k.Hosts = hosts
	k.Comment = comment

	if pubKey != nil {
		k.SHA256Fingerprint = ssh.FingerprintSHA256(pubKey)
		k.MD5Fingerprint = ssh.FingerprintLegacyMD5(pubKey)
	}

	return k
}

func fromDomainKnownhost(k *domain.KnownHost) knownhost {
	return knownhost{
		UUID:  k.UUID,
		Entry: k.Entry,
	}
}

func SaveKnownHost(k *domain.KnownHost) error {
	tx := db.Begin()

	gK := fromDomainKnownhost(k)

	tx.
		Where(knownhost{Entry: k.Entry}).
		Assign(&gK).
		FirstOrCreate(&gK)

	return tx.Commit().Error
}

func DeleteKnownHost(k *domain.KnownHost) error {
	tx := db.Begin()

	gK := fromDomainKnownhost(k)

	if err := tx.Delete(gK).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func GetKnownHosts() (r []*domain.KnownHost, t int) {
	t = 0

	gKs := []knownhost{}
	db.Find(&gKs).Count(&t)

	for _, gK := range gKs {
		r = append(r, gK.toDomainKnownhost())
	}

	return r, t
}
