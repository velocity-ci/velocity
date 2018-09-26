package velocity

import (
	"io"
	"strings"
	"time"

	"github.com/go-cmd/cmd"
	"go.uber.org/zap"
)

type App interface {
	Start()
	Stop() error
}

func runCmd(writer io.Writer, shCmd []string, env []string) cmd.Status {
	opts := cmd.Options{Buffered: false, Streaming: true}
	c := cmd.NewCmdOptions(opts, shCmd[0], shCmd[1:len(shCmd)]...)
	c.Env = env
	stdout := []string{}
	stderr := []string{}
	go func() {
		for line := range c.Stdout {
			writer.Write([]byte(line))
			stdout = append(stdout, line)
		}
	}()
	go func() {
		for line := range c.Stderr {
			writer.Write([]byte(line))
			stderr = append(stderr, line)
		}
	}()

	GetLogger().Debug("running command", zap.Strings("cmd", shCmd))
	go func() {
		<-time.After(5 * time.Second)
		if !c.Status().Complete && (len(stdout) < 1 && len(stderr) < 1) {
			GetLogger().Debug("5s", zap.Strings("cmd", shCmd), zap.Strings("stdout", stdout), zap.Strings("stderr", stderr), zap.Int("status", c.Status().Exit))
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

	GetLogger().Debug("completed cmd",
		zap.String("cmd", strings.Join(shCmd, " ")),
		zap.Int("exited", finalStatus.Exit),
		zap.Strings("stdout", finalStatus.Stdout),
		zap.Strings("stderr", finalStatus.Stderr),
		zap.Float64("runtime (s)", finalStatus.Runtime),
	)

	return finalStatus
}
