package knownhost

import (
	"log"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
)

type Repository interface {
	Create(k KnownHost) KnownHost
	Delete(k KnownHost)
	GetByID(id string) (KnownHost, error)
	GetAll(q KnownHostQuery) ([]KnownHost, uint64)
	GetAllEntries() []string
}

type KnownHostQuery struct {
	Amount uint64
	Page   uint64
}

type KnownHost struct {
	ID                string   `json:"id"`
	Entry             string   `json:"entry"`
	Hosts             []string `json:"hosts"`
	Comment           string   `json:"comment"`
	SHA256Fingerprint string   `json:"sha256"`
	MD5Fingerprint    string   `json:"md5"`
}

type ResponseKnownHost struct {
	ID                string   `json:"id"`
	Hosts             []string `json:"hosts"`
	Comment           string   `json:"comment"`
	SHA256Fingerprint string   `json:"sha256"`
	MD5Fingerprint    string   `json:"md5"`
}

type RequestKnownHost struct {
	Entry string `json:"entry" validate:"required,knownHostValid,knownHostUnique,min=10"`
}

func NewKnownHost(e string) KnownHost {
	_, hosts, pubKey, comment, _, err := ssh.ParseKnownHosts([]byte(e))
	if err != nil {
		log.Fatal(err)
	}

	id := uuid.NewV1().String()

	if pubKey == nil {
		return KnownHost{
			ID:      id,
			Hosts:   hosts,
			Comment: comment,
		}
	}

	return KnownHost{
		Entry:             e,
		ID:                id,
		Hosts:             hosts,
		Comment:           comment,
		SHA256Fingerprint: ssh.FingerprintSHA256(pubKey),
		MD5Fingerprint:    ssh.FingerprintLegacyMD5(pubKey),
	}
}

func NewResponseKnownHost(k KnownHost) ResponseKnownHost {
	return ResponseKnownHost{
		ID:                k.ID,
		Hosts:             k.Hosts,
		Comment:           k.Comment,
		SHA256Fingerprint: k.SHA256Fingerprint,
		MD5Fingerprint:    k.MD5Fingerprint,
	}
}

type ManyResponseKnownHost struct {
	Total  uint64              `json:"total"`
	Result []ResponseKnownHost `json:"result"`
}
