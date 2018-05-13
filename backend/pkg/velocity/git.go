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

	"golang.org/x/crypto/ssh/agent"

	"github.com/Sirupsen/logrus"
	"github.com/go-cmd/cmd"
	"github.com/gosimple/slug"
	"golang.org/x/crypto/ssh"
)

type SSHKeyError string

func (s SSHKeyError) Error() string {
	return string(s)
}

type GitRepository struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}

// TODO:
// Clone
// x Clone Depth (1)
// - SSH authentication with just private key
// x Recurse submodules (delay support)
// Sync
// x Get remote branches
// x Get commits at HEAD of each branch
// x Checkout commit by sha

func Clone(
	r *GitRepository,
	bare,
	full,
	submodule bool,
	writer io.Writer,
) (*RawRepository, error) {
	psuedoRandom := rand.NewSource(time.Now().UnixNano())
	randNumber := rand.New(psuedoRandom)
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	dir := fmt.Sprintf("%s/workspaces/_%s-%d", wd, slug.Make(r.Address), randNumber.Int63())
	os.RemoveAll(dir)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		logrus.Fatal(err)
		return nil, err
	}

	shCmd := []string{"git", "clone", "--progress"}

	if bare {
		shCmd = append(shCmd, "--bare")
	}

	if !full {
		shCmd = append(shCmd, "--depth=1")
	}

	if submodule {
		shCmd = append(shCmd, "--recurse-submodules")
	}

	shCmd = append(shCmd, r.Address)
	shCmd = append(shCmd, dir)

	if r.Address[:3] == "git" {
		key, err := ssh.ParseRawPrivateKey([]byte(r.PrivateKey))
		if err != nil {
			return nil, err
		}
		if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
			a := agent.NewClient(sshAgent)
			a.Add(agent.AddedKey{PrivateKey: key})
			signer, _ := ssh.NewSignerFromKey(key)

			defer a.Remove(signer.PublicKey())
		}
	}

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
			logrus.Infof("skipped branch: %s", line)
			continue
		}
		logrus.Infof("got branch: %s", line)
		b = append(b, line)
	}

	return b
}

func (r *RawRepository) GetCommitAtHeadOfBranch(branch string) *RawCommit {
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
		logrus.Fatalf("could not get work dir: %v", err)
	}
	r.pwd = cwd
	os.Chdir(r.Directory)
}

func (r *RawRepository) done() {
	os.Chdir(r.pwd)
	r.pwd = ""
	r.RUnlock()
}

func (r *RawRepository) GetCommitInfo(sha string) *RawCommit {
	r.init()
	defer r.done()
	shCmd := []string{"git", "show", "-s", `--format=%H%n%aI%n%aE%n%aN%n%G?%n%s`, sha}
	c := cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	s := <-c.Start()

	authorDate, _ := time.Parse(time.RFC3339, strings.TrimSpace(s.Stdout[1]))

	return &RawCommit{
		SHA:         strings.TrimSpace(s.Stdout[0]),
		AuthorDate:  authorDate,
		AuthorEmail: strings.TrimSpace(s.Stdout[2]),
		AuthorName:  strings.TrimSpace(s.Stdout[3]),
		Signed:      strings.TrimSpace(s.Stdout[4]),
		Message:     strings.TrimSpace(s.Stdout[5]),
	}
}

func (r *RawRepository) GetCurrentCommitInfo() *RawCommit {
	r.init()
	defer r.done()
	shCmd := []string{"git", "rev-parse", "HEAD"}
	c := cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	s := <-c.Start()

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
		return s.Error
	}

	if s.Exit != 0 {
		logrus.Error(s.Stdout, s.Stderr)
		return fmt.Errorf("non-zero exit in clone: %d", s.Exit)
	}

	return nil
}

func (r *RawRepository) Checkout(sha string) error {
	r.init()
	defer r.done()

	if err := r.Clean(); err != nil {
		return err
	}

	shCmd := []string{"git", "checkout", "--force", sha}
	c := cmd.NewCmd(shCmd[0], shCmd[1:len(shCmd)]...)
	s := <-c.Start()

	if err := handleStatusError(s); err != nil {
		return err
	}

	return nil
}
