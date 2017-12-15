package main

import (
	"github.com/velocity-ci/velocity/backend/api/domain/knownhost"
	"github.com/velocity-ci/velocity/backend/api/slave"
)

func updateKnownHosts(c *slave.KnownHostCommand) {
	fM := knownhost.NewFileManager()
	fM.Clear()
	for _, k := range c.KnownHosts {
		fM.Save(k)
	}
}
