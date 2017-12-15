package knownhost

import (
	"log"

	"golang.org/x/crypto/ssh"
)

type Repository interface {
	Create(k KnownHost) KnownHost
	Delete(k KnownHost)
	GetByID(id string) (KnownHost, error)
	GetAll(q KnownHostQuery) ([]KnownHost, uint64)
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

func NewResponseKnownHost(k KnownHost) ResponseKnownHost {
	_, hosts, pubKey, comment, _, err := ssh.ParseKnownHosts([]byte(k.Entry))
	if err != nil {
		log.Fatal(err)
	}

	if pubKey == nil {
		return &ResponseKnownHost{
			Hosts:   hosts,
			Comment: comment,
		}
	}

	return &ResponseKnownHost{
		Hosts:             hosts,
		Comment:           comment,
		SHA256Fingerprint: ssh.FingerprintSHA256(pubKey),
		MD5Fingerprint:    ssh.FingerprintLegacyMD5(pubKey),
	}
}
