package git

import (
	"strings"

	"github.com/velocity-ci/velocity/backend/pkg/exec"
)

func (r *RawRepository) GetDescribe() string {
	r.RLock()
	defer r.RUnlock()
	shCmd := []string{"git", "describe", "--always"}
	s := exec.Run(shCmd, r.Directory, []string{}, nil)

	return strings.TrimSpace(s.Stdout[0])
}
