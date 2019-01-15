package git

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/exec"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

type RawCommit struct {
	SHA         string
	AuthorDate  time.Time
	AuthorEmail string
	AuthorName  string
	Signed      string
	Message     string
}

func (r *RawRepository) GetCommitInfo(sha string) (*RawCommit, error) {
	r.RLock()
	defer r.RUnlock()
	shCmd := []string{"git", "show", "-s", `--format=%H%n%aI%n%aE%n%aN%n%GK%n%s`, sha}
	s := exec.Run(shCmd, r.Directory, []string{}, nil)

	if len(s.Stdout) < 6 {
		velocity.GetLogger().Error("unexpected commit info output", zap.Strings("stdout", s.Stdout), zap.Strings("stderr", s.Stderr))
		return nil, fmt.Errorf("unexpected commit info output")
	}

	authorDate, _ := time.Parse(time.RFC3339, strings.TrimSpace(s.Stdout[1]))

	return &RawCommit{
		SHA:         strings.TrimSpace(s.Stdout[0]),
		AuthorDate:  authorDate,
		AuthorEmail: strings.TrimSpace(s.Stdout[2]),
		AuthorName:  strings.TrimSpace(s.Stdout[3]),
		Signed:      strings.TrimSpace(s.Stdout[4]),
		Message:     strings.TrimSpace(s.Stdout[5]),
	}, nil
}

func (r *RawRepository) GetCurrentCommitInfo() (*RawCommit, error) {
	r.RLock()
	defer r.RUnlock()
	shCmd := []string{"git", "rev-parse", "HEAD"}
	s := exec.Run(shCmd, r.Directory, []string{}, nil)

	// GetLogger().Debug("git rev-parse HEAD", zap.Strings("stdout", s.Stdout), zap.Strings("stderr", s.Stderr))

	return r.GetCommitInfo(strings.TrimSpace(s.Stdout[0]))
}

func (r *RawRepository) GetTotalCommits() uint64 {
	r.RLock()
	defer r.RUnlock()
	shCmd := []string{"git", "rev-list", "--all", "--count"}
	// writer := &BlankWriter{}
	s := exec.Run(shCmd, r.Directory, []string{}, nil)

	total, err := strconv.ParseUint(strings.TrimSpace(s.Stdout[0]), 10, 64)
	if err != nil {
		velocity.GetLogger().Error("could not get total commits", zap.Error(err))
	}

	return total
}
