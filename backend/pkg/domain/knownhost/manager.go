package knownhost

import (
	"github.com/asdine/storm"
	"github.com/go-playground/universal-translator"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"golang.org/x/crypto/ssh"
	govalidator "gopkg.in/go-playground/validator.v9"
)

type Manager struct {
	validator *validator
	db        *stormDB
}

func NewManager(
	db *storm.DB,
	validator *govalidator.Validate,
	translator ut.Translator,
) *Manager {
	m := &Manager{
		db: newStormDB(db),
	}
	m.validator = newValidator(validator, translator, m)
	return m
}

func (m *Manager) Create(entry string) (*KnownHost, *domain.ValidationErrors) {
	k := &KnownHost{
		Entry: entry,
	}

	if err := m.validator.Validate(k); err != nil {
		return nil, err
	}

	_, hosts, pubKey, comment, _, _ := ssh.ParseKnownHosts([]byte(entry))

	k.ID = uuid.NewV1().String()
	k.Hosts = hosts
	k.Comment = comment

	if pubKey != nil {
		k.SHA256Fingerprint = ssh.FingerprintSHA256(pubKey)
		k.MD5Fingerprint = ssh.FingerprintLegacyMD5(pubKey)
	}

	m.db.save(k)

	return k, nil
}

func (m *Manager) Update(k *KnownHost) error {
	return m.db.save(k)
}

func (m *Manager) Delete(k *KnownHost) error {
	if err := m.db.delete(k); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Exists(entry string) bool {
	return m.db.exists(entry)
}

func (m *Manager) GetAll(q *domain.PagingQuery) ([]*KnownHost, int) {
	return m.db.getAll(q)
}
