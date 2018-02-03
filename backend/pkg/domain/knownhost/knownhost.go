package knownhost

type KnownHost struct {
	UUID              string   `json:"id"`
	Entry             string   `json:"entry" validate:"required,knownHostValid"`
	Hosts             []string `json:"hosts"`
	Comment           string   `json:"comment"`
	SHA256Fingerprint string   `json:"sha256"`
	MD5Fingerprint    string   `json:"md5"`
}
