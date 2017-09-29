package domain

import "golang.org/x/crypto/ssh"

type KnownHost struct {
	Entry string `json:"entry"`
}

type ResponseKnownHost struct {
	Hosts             []string `json:"hosts"`
	Comment           string   `json:"comment"`
	SHA256Fingerprint string   `json:"sha256"`
	MD5Fingerprint    string   `json:"md5"`
}

func (k *KnownHost) ToResponseKnownHost() *ResponseKnownHost {
	_, hosts, pubKey, comment, _, _ := ssh.ParseKnownHosts([]byte(k.Entry))

	return &ResponseKnownHost{
		Hosts:             hosts,
		Comment:           comment,
		SHA256Fingerprint: ssh.FingerprintSHA256(pubKey),
		MD5Fingerprint:    ssh.FingerprintLegacyMD5(pubKey),
	}
}
