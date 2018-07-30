package velocity

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"golang.org/x/crypto/ssh/agent"

	"github.com/go-cmd/cmd"
	"github.com/gosimple/slug"
	"golang.org/x/crypto/ssh"
)

type SSHKeyError string

func (s SSHKeyError) Error() string {
	return string(s)
}

type HostKeyError string

func (s HostKeyError) Error() string {
	return string(s)
}

type GitRepository struct {
	Address    string        `json:"address"`
	PrivateKey string        `json:"privateKey"`
	PublicKey  ssh.PublicKey `json:"-"`
	Agent      agent.Agent   `json:"-"`
}

const WorkspaceDir = "/opt/velocityci/workspaces"

func getUniqueWorkspace(r *GitRepository) (string, error) {
	psuedoRandom := rand.NewSource(time.Now().UnixNano())
	randNumber := rand.New(psuedoRandom)
	dir := fmt.Sprintf("%s/_%s-%d", WorkspaceDir, slug.Make(r.Address), randNumber.Int63())
	os.RemoveAll(dir)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		GetLogger().Fatal("could not create unique workspace", zap.Error(err))
		return "", err
	}

	return dir, nil
}

func handleGitSSH(r *GitRepository) (ssh.Signer, agent.Agent, error) {
	key, err := ssh.ParseRawPrivateKey([]byte(r.PrivateKey))
	if err != nil {
		return nil, nil, err.(SSHKeyError)
	}
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		a := agent.NewClient(sshAgent)
		a.Add(agent.AddedKey{PrivateKey: key})
		signer, _ := ssh.NewSignerFromKey(key)
		GetLogger().Debug("added ssh key to ssh-agent", zap.String("address", r.Address))
		return signer, a, nil
	}

	return nil, nil, fmt.Errorf("ssh-agent not found")
}

func initWorkspace(r *GitRepository) (string, error) {
	dir, _ := getUniqueWorkspace(r)
	os.Chdir(dir)
	shCmd := []string{"git", "init"}
	c := cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	<-c.Start()

	shCmd = []string{"git", "remote", "add", "origin", r.Address}
	c = cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	<-c.Start()

	if r.Address[:3] == "git" {
		signer, agent, err := handleGitSSH(r)
		if err != nil {
			return "", err
		}
		r.Agent = agent
		r.PublicKey = signer.PublicKey()
	}

	return dir, nil
}

func cleanSSHAgent(r *GitRepository) {
	if r.Agent != nil {
		r.Agent.Remove(r.PublicKey)
		GetLogger().Debug("removed ssh key from ssh-agent", zap.String("address", r.Address))
	}
}

func Validate(r *GitRepository) (bool, error) {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)

	dir, err := initWorkspace(r)
	if err != nil {
		return false, err
	}
	defer cleanSSHAgent(r)
	os.Chdir(dir)

	shCmd := []string{"git", "ls-remote"}
	c := cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	s := <-c.Start()

	if s.Exit != 0 {
		err := fmt.Errorf(strings.Join(s.Stderr, " "))
		if strings.Contains(err.Error(), "Host key verification failed") {
			err = HostKeyError(err.Error())
		}
		return false, err
	}

	return true, nil
}

type CloneOptions struct {
	Bare      bool
	Full      bool
	Submodule bool
	Commit    string
}

func Clone(
	r *GitRepository,
	writer io.Writer,
	cloneOpts *CloneOptions,
) (*RawRepository, error) {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)

	dir, err := initWorkspace(r)
	if err != nil {
		return nil, err
	}
	defer cleanSSHAgent(r)
	os.Chdir(dir)

	shCmd := []string{"git", "fetch", "--progress"}

	if cloneOpts.Bare {
		shCmd = append(shCmd, "--bare")
	}

	if !cloneOpts.Full {
		shCmd = append(shCmd, "--depth=1")
	}

	if cloneOpts.Submodule {
		shCmd = append(shCmd, "--recurse-submodules")
	}

	if len(cloneOpts.Commit) > 0 {
		// shCmd = append(shCmd, "origin", cloneOpts.Commit)
	}

	GetLogger().Info("cloning repository", zap.String("cmd", strings.Join(shCmd, " ")))

	opts := cmd.Options{Buffered: false, Streaming: true}
	c := cmd.NewCmdOptions(opts, shCmd[0], shCmd[1:len(shCmd)]...)
	stdOut := []string{}
	stdErr := []string{}
	go func() {
		for {
			select {
			case line := <-c.Stdout:
				writer.Write([]byte(line))
				stdOut = append(stdOut, line)
			case line := <-c.Stderr:
				writer.Write([]byte(line))
				stdErr = append(stdErr, line)
			}
		}
	}()
	s := <-c.Start()
	for len(c.Stdout) > 0 || len(c.Stderr) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
	s.Stdout = stdOut
	s.Stderr = stdErr

	if err := handleStatusError(s); err != nil {
		os.RemoveAll(dir)
		return nil, err
	}

	GetLogger().Info("cloned repository", zap.String("address", r.Address))

	return &RawRepository{Directory: dir}, nil
}

func (r *RawRepository) GetBranches() (b []string) {
	r.init()
	defer r.done()

	shCmd := []string{"git", "branch", "--remote"}
	c := cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	s := <-c.Start()
	for _, line := range s.Stdout {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "origin/")
		if strings.HasPrefix(line, "HEAD") {
			continue
		}
		b = append(b, line)
	}

	return b
}

func (r *RawRepository) GetCommitAtHeadOfBranch(branch string) (*RawCommit, error) {
	r.init()
	defer r.done()
	commitSha := r.RevParse(fmt.Sprintf("origin/%s", branch))

	return r.GetCommitInfo(commitSha)
}

func (r *RawRepository) RevParse(obj string) string {
	r.init()
	defer r.done()
	shCmd := []string{"git", "rev-parse", obj}
	c := cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	s := <-c.Start()

	return strings.TrimSpace(s.Stdout[0])
}

type RawCommit struct {
	SHA         string
	AuthorDate  time.Time
	AuthorEmail string
	AuthorName  string
	Signed      string
	Message     string
}

type RawRepository struct {
	Directory string
	pwd       string
	sync.RWMutex
}

func (r *RawRepository) init() {
	r.RLock()
	cwd, err := os.Getwd()
	if err != nil {
		GetLogger().Fatal("could not get working directory", zap.Error(err))
	}
	r.pwd = cwd
	os.Chdir(r.Directory)
}

func (r *RawRepository) done() {
	os.Chdir(r.pwd)
	r.pwd = ""
	r.RUnlock()
}

func (r *RawRepository) GetCommitInfo(sha string) (*RawCommit, error) {
	r.init()
	defer r.done()
	shCmd := []string{"git", "show", "-s", `--format=%H%n%aI%n%aE%n%aN%n%GK%n%s`, sha}
	c := cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	s := <-c.Start()

	if len(s.Stdout) < 6 {
		GetLogger().Error("unexpected commit info output", zap.Strings("stdout", s.Stdout), zap.Strings("stderr", s.Stderr))
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
	r.init()
	defer r.done()
	shCmd := []string{"git", "rev-parse", "HEAD"}
	c := cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	s := <-c.Start()

	GetLogger().Debug("git rev-parse HEAD", zap.Strings("stdout", s.Stdout), zap.Strings("stderr", s.Stderr))

	return r.GetCommitInfo(strings.TrimSpace(s.Stdout[0]))
}

func (r *RawRepository) GetDescribe() string {
	r.init()
	defer r.done()
	shCmd := []string{"git", "describe", "--always"}
	c := cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	s := <-c.Start()

	return strings.TrimSpace(s.Stdout[0])
}

func (r *RawRepository) Clean() error {
	r.init()
	defer r.done()

	shCmd := []string{"git", "clean", "-fd"}
	c := cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	s := <-c.Start()

	return handleStatusError(s)
}

func handleStatusError(s cmd.Status) error {
	if s.Error != nil {
		GetLogger().Error("unknown cmd error", zap.Error(s.Error))
		return s.Error
	}

	if s.Exit != 0 {
		GetLogger().Error("non-zero exit in git", zap.Strings("stdout", s.Stdout), zap.Strings("stderr", s.Stderr))
		return fmt.Errorf(strings.Join(s.Stderr, "\n"))
	}

	return nil
}

func (r *RawRepository) Checkout(ref string) error {
	r.init()
	defer r.done()

	if err := r.Clean(); err != nil {
		return err
	}

	shCmd := []string{"git", "checkout", "--force", ref}
	c := cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	s := <-c.Start()

	if err := handleStatusError(s); err != nil {
		return err
	}

	GetLogger().Debug("checked out", zap.String("reference", ref), zap.String("repository", r.Directory))

	return nil
}

func (r *RawRepository) GetDefaultBranch() string {
	r.init()
	defer r.done()
	shCmd := []string{"git", "remote", "show", "origin"}
	c := cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	s := <-c.Start()

	defaultBranch := strings.TrimSpace(strings.Split(s.Stdout[3], ":")[1])

	return defaultBranch
}
