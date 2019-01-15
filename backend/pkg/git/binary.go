package git

import (
	"strings"

	"github.com/velocity-ci/velocity/backend/pkg/exec"
)

func GetVersion() string {
	shCmd := []string{"git", "--version"}
	s := exec.Run(shCmd, "", []string{}, nil)

	return strings.TrimSpace(s.Stdout[0])[12:]
}
