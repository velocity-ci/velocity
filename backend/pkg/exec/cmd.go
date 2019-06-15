package exec

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"go.uber.org/zap"
	"golang.org/x/net/http/httpproxy"
)

func Run(shCmd []string, directory string, env []string, writer io.Writer) cmd.Status {
	opts := cmd.Options{Buffered: false, Streaming: true}
	c := cmd.NewCmdOptions(opts, shCmd[0], shCmd[1:]...)
	c.Env = respectProxyEnv(env)
	c.Dir = directory
	stdout := []string{}
	stderr := []string{}
	go func() {
		for line := range c.Stdout {
			if writer != nil {
				writer.Write([]byte(line))
			}
			stdout = append(stdout, line)
		}
	}()
	go func() {
		for line := range c.Stderr {
			if writer != nil {
				writer.Write([]byte(line))
			}
			stderr = append(stderr, line)
		}
	}()

	logging.GetLogger().Debug("running command", zap.Strings("cmd", shCmd), zap.String("directory", directory))
	go func() {
		<-time.After(5 * time.Second)
		if !c.Status().Complete && (len(stdout) < 1 && len(stderr) < 1) {
			logging.GetLogger().Debug("5s", zap.Strings("cmd", shCmd), zap.Strings("stdout", stdout), zap.Strings("stderr", stderr), zap.Int("status", c.Status().Exit))
			c.Stop()
		}
	}()
	s := c.Start()

	finalStatus := <-s
	close(c.Stdout)
	close(c.Stderr)
	time.Sleep(10 * time.Millisecond)
	finalStatus.Stdout = stdout
	finalStatus.Stderr = stderr

	logging.GetLogger().Debug("completed cmd",
		zap.String("cmd", strings.Join(shCmd, " ")),
		zap.Int("exited", finalStatus.Exit),
		zap.Strings("stdout", finalStatus.Stdout),
		zap.Strings("stderr", finalStatus.Stderr),
		zap.Float64("runtime (s)", finalStatus.Runtime),
	)

	return finalStatus
}

func respectProxyEnv(env []string) []string {
	config := httpproxy.FromEnvironment()
	if len(config.HTTPProxy) > 1 {
		env = append(env, fmt.Sprintf("HTTP_PROXY=%s", config.HTTPProxy))
		env = append(env, fmt.Sprintf("http_proxy=%s", config.HTTPProxy))
	}
	if len(config.HTTPSProxy) > 1 {
		env = append(env, fmt.Sprintf("HTTPS_PROXY=%s", config.HTTPSProxy))
		env = append(env, fmt.Sprintf("https_proxy=%s", config.HTTPSProxy))
	}
	if len(config.NoProxy) > 1 {
		env = append(env, fmt.Sprintf("NO_PROXY=%s", config.NoProxy))
		env = append(env, fmt.Sprintf("no_proxy=%s", config.NoProxy))
	}

	return env
}

func GetStatusError(s cmd.Status) error {
	if s.Error != nil {
		logging.GetLogger().Error("unknown cmd error", zap.Error(s.Error))
		return s.Error
	}

	if s.Exit != 0 {
		logging.GetLogger().Error("non-zero exit code", zap.Strings("stdout", s.Stdout), zap.Strings("stderr", s.Stderr))
		return fmt.Errorf(strings.Join(s.Stderr, "\n"))
	}

	return nil
}
