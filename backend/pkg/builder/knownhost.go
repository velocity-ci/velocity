package builder

import (
	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"
)

func updateKnownHosts(c *builder.KnownHostCtrl) {
	fM := knownhost.NewFileManager("")
	fM.WriteAll(c.KnownHosts)
}
