package builder

import (
	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"
)

func (b *Builder) updateKnownHosts(c *KnownHostPayload) {
	fM := knownhost.NewFileManager("")
	fM.WriteAll(c.KnownHosts)
}

type KnownHostPayload struct {
	KnownHosts []*knownhost.KnownHost `json:"knownHosts"`
}
