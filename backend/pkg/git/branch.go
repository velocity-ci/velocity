package git

import (
	"fmt"
	"strings"

	"github.com/velocity-ci/velocity/backend/pkg/exec"
)

func CurrentBranch(projectRoot string) (string, error) {
	shCmd := []string{"git", "status"}
	s := exec.Run(shCmd, projectRoot, []string{}, nil)

	if err := exec.GetStatusError(s); err != nil {
		return "", err
	}

	parts := strings.Split(s.Stdout[0], " ")
	if parts[0] == "On" {
		return parts[2], nil
	}
	return "", fmt.Errorf(s.Stdout[0])
}

func (r *RawRepository) GetDefaultBranch() (string, error) {
	r.RLock()
	defer r.RUnlock()

	deferFunc, err := handleGitSSH(r.Repository)
	if err != nil {
		return "", err
	}
	defer deferFunc(r.Repository)

	shCmd := []string{"git", "remote", "show", "origin"}
	s := exec.Run(shCmd, r.Directory, []string{}, nil)

	if err := exec.GetStatusError(s); err != nil {
		return "", err
	}

	defaultBranch := strings.TrimSpace(strings.Split(s.Stdout[3], ":")[1])

	return defaultBranch, nil
}
