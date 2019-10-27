package main

import (
	"fmt"
	"os"
	"time"

	"github.com/velocity-ci/velocity/backend/cmd/vcli/cmds"
	"github.com/velocity-ci/velocity/backend/pkg/exec"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
)

func main() {
	defer exec.Time(time.Now(), "main")
	if err := cmds.Execute(); err != nil {
		fmt.Fprintf(os.Stdout, output.ColorFmt(output.ANSIError, "%s", "\n"), err)
		os.Exit(1)
	}
}
