package git

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/velocity-ci/velocity/backend/pkg/auth"
	"github.com/velocity-ci/velocity/backend/pkg/exec"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Repository struct {
	Address    string        `json:"address"`
	PrivateKey string        `json:"privateKey"`
	PublicKey  ssh.PublicKey `json:"-"`
	Agent      agent.Agent   `json:"-"`
}

type RawRepository struct {
	Repository *Repository
	Directory  string
	sync.RWMutex
}

type CloneOptions struct {
	Bare      bool
	Depth     uint64
	Submodule bool
	Commit    string
}

type SSHKeyError string

func (s SSHKeyError) Error() string {
	return string(s)
}

type HostKeyError string

func (s HostKeyError) Error() string {
	return string(s)
}

func getUniqueWorkspace(r *Repository) (string, error) {
	dir := fmt.Sprintf("%s/_%s-%s",
		"",
		// slug.Make(r.Address),
		auth.RandomString(8),
		auth.RandomString(8),
	)

	err := os.RemoveAll(dir)
	if err != nil {
		velocity.GetLogger().Fatal("could not create unique workspace", zap.Error(err))
		return "", err
	}
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		velocity.GetLogger().Fatal("could not create unique workspace", zap.Error(err))
		return "", err
	}

	return dir, nil
}

func initWorkspace(r *Repository, dir string, writer io.Writer) (string, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		err := os.RemoveAll(dir)
		if err != nil {
			return "", err
		}
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return "", err
		}
		shCmd := []string{"git", "init"}
		exec.Run(shCmd, dir, []string{}, nil)
	}

	shCmd := []string{"git", "remote", "--verbose"}
	cmdRes := exec.Run(shCmd, dir, []string{}, nil)
	// stdErr := strings.Join(cmdRes.Stderr, "\n")
	stdOut := strings.Join(cmdRes.Stdout, "\n")
	fetch := strings.Contains(stdOut, fmt.Sprintf("origin	%s (fetch)", r.Address))
	push := strings.Contains(stdOut, fmt.Sprintf("origin	%s (push)", r.Address))
	if !fetch || !push {
		shCmd = []string{"git", "remote", "remove", "origin"}
		exec.Run(shCmd, dir, []string{}, nil)
		shCmd = []string{"git", "remote", "add", "origin", r.Address}
		exec.Run(shCmd, dir, []string{}, nil)
	}

	return dir, nil
}

func handleGitSSH(r *Repository) (d func(*Repository), err error) {
	d = func(*Repository) {}
	if r.Address[:3] == "git" {
		key, err := ssh.ParseRawPrivateKey([]byte(r.PrivateKey))
		if err != nil {
			return d, err.(SSHKeyError)
		}

		if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
			a := agent.NewClient(sshAgent)
			a.Add(agent.AddedKey{PrivateKey: key})
			signer, _ := ssh.NewSignerFromKey(key)
			velocity.GetLogger().Debug("added ssh key to ssh-agent", zap.String("address", r.Address))
			r.Agent = a
			r.PublicKey = signer.PublicKey()
			return cleanSSHAgent, nil
		}

		return d, fmt.Errorf("ssh-agent not found")
	}

	return d, nil
}

func cleanSSHAgent(r *Repository) {
	if r.Agent != nil {
		r.Agent.Remove(r.PublicKey)
		velocity.GetLogger().Debug("removed ssh key from ssh-agent", zap.String("address", r.Address))
	}
}

func Clone(
	r *Repository,
	cloneOpts *CloneOptions,
	directory string,
	writer io.Writer,
) (*RawRepository, error) {
	// wd, _ := os.Getwd()
	// defer os.Chdir(wd)

	dir, err := initWorkspace(r, directory, writer)
	if err != nil {
		return nil, err
	}
	// os.Chdir(dir)

	shCmd := []string{"git", "fetch", "--progress"}

	if cloneOpts.Bare {
		shCmd = append(shCmd, "--bare")
	}

	if cloneOpts.Depth > 0 {
		shCmd = append(shCmd, fmt.Sprintf("--depth=%d", cloneOpts.Depth))
	}

	if cloneOpts.Submodule {
		shCmd = append(shCmd, "--recurse-submodules")
	}

	if len(cloneOpts.Commit) > 0 {
		// shCmd = append(shCmd, "origin", cloneOpts.Commit)
	}

	velocity.GetLogger().Info("fetching repository", zap.String("cmd", strings.Join(shCmd, " ")))

	deferFunc, err := handleGitSSH(r)
	if err != nil {
		return nil, err
	}
	defer deferFunc(r)

	s := exec.Run(shCmd, dir, []string{}, writer)
	if err := exec.GetStatusError(s); err != nil {
		os.RemoveAll(dir)
		return nil, err
	}

	velocity.GetLogger().Info("fetched repository", zap.String("address", r.Address), zap.String("dir", dir))

	return &RawRepository{Directory: dir, Repository: r}, nil
}

func (r *RawRepository) Clean() error {
	r.RLock()
	defer r.RUnlock()

	shCmd := []string{"git", "clean", "-fd"}
	s := exec.Run(shCmd, r.Directory, []string{}, nil)

	return exec.GetStatusError(s)
}

func (r *RawRepository) Checkout(ref string) error {
	r.RLock()
	defer r.RUnlock()

	if err := r.Clean(); err != nil {
		return err
	}

	shCmd := []string{"git", "checkout", "--force", ref}
	s := exec.Run(shCmd, r.Directory, []string{}, nil)

	if err := exec.GetStatusError(s); err != nil {
		return err
	}

	velocity.GetLogger().Debug("checked out", zap.String("reference", ref), zap.String("repository", r.Directory))

	return nil
}

func (r *RawRepository) IsDirty() bool {
	r.RLock()
	defer r.RUnlock()

	shCmd := []string{"git", "status", "--porcelain"}
	s := exec.Run(shCmd, r.Directory, []string{}, nil)

	if len(s.Stdout) > 0 {
		return true
	}

	return false
}
