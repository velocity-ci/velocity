package knownhost

import (
	"fmt"

	"github.com/asdine/storm"
	"github.com/go-playground/universal-translator"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"golang.org/x/crypto/ssh"
	govalidator "gopkg.in/go-playground/validator.v9"
)

// Event constants
const (
	EventCreate = "knownhost:new"
	EventDelete = "knownhost:delete"
)

type Manager struct {
	validator   *validator
	db          *stormDB
	fileManager *FileManager
	brokers     []domain.Broker
}

func NewManager(
	db *storm.DB,
	validator *govalidator.Validate,
	translator ut.Translator,
	homedir string,
) *Manager {
	m := &Manager{
		db:          newStormDB(db),
		brokers:     []domain.Broker{},
		fileManager: NewFileManager(homedir),
	}
	m.validator = newValidator(validator, translator, m)
	knownHosts, _ := m.GetAll(&domain.PagingQuery{Limit: 50, Page: 1})
	m.fileManager.WriteAll(knownHosts)
	return m
}

func (m *Manager) AddBroker(b domain.Broker) {
	m.brokers = append(m.brokers, b)
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
	kH, _ := m.GetAll(&domain.PagingQuery{Limit: 100})
	m.fileManager.WriteAll(kH)

	for _, b := range m.brokers {
		b.EmitAll(&domain.Emit{
			Topic:   "knownhosts",
			Event:   EventCreate,
			Payload: k,
		})
	}

	return k, nil
}

func (m *Manager) Delete(k *KnownHost) error {
	if err := m.db.delete(k); err != nil {
		return err
	}
	kH, _ := m.GetAll(&domain.PagingQuery{Limit: 100})
	m.fileManager.WriteAll(kH)
	for _, b := range m.brokers {
		b.EmitAll(&domain.Emit{
			Topic:   fmt.Sprintf("knownhosts:%s", k.ID),
			Event:   EventDelete,
			Payload: k,
		})
	}
	return nil
}

func (m *Manager) Exists(entry string) bool {
	return m.db.exists(entry)
}

func (m *Manager) GetAll(q *domain.PagingQuery) ([]*KnownHost, int) {
	return m.db.getAll(q)
}
