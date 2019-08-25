package docker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"go.uber.org/zap"
)

func HandleOutput(body io.ReadCloser, censored []string, writer io.Writer) {
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		allBytes := scanner.Bytes()
		o := ""
		if strings.Contains(string(allBytes), "status") {
			o = handlePullPushOutput(allBytes)
		} else if strings.Contains(string(allBytes), "stream") {
			o = handleBuildOutput(allBytes)
		} else if strings.Contains(string(allBytes), "progressDetail") {
			o = "*"
		} else {
			o = handleLogOutput(allBytes)
		}
		if o != "*" {
			for _, c := range censored {
				o = strings.Replace(o, c, "***", -1)
			}
			writer.Write([]byte(o))
		}
	}
	body.Close()
}

func handleLogOutput(b []byte) string {
	if len(b) <= 8 {
		return ""
	}
	line := string(b[8:])
	if !strings.Contains(line, "\r") {
		return fmt.Sprintf("%s\n", line)
	}
	return line
}

var imageIDProgress = map[string]string{}

func handlePullPushOutput(b []byte) string {
	type pullOutput struct {
		Status   string `json:"status"`
		Progress string `json:"progress"`
		ID       string `json:"id"`
	}
	var o pullOutput
	json.Unmarshal(b, &o)

	s := ""
	if len(o.ID) > 0 {
		s += fmt.Sprintf("%s: ", o.ID)
	}
	if len(o.Progress) > 0 {
		s += o.Progress
	} else {
		s += o.Status
	}
	// add padding to 80
	for len(s) < 100 {
		s += " "
	}
	if strings.Contains(o.Status, "Downloaded newer image") ||
		strings.Contains(o.Status, "Pulling from") ||
		strings.Contains(o.Status, "Image is up to date") ||
		strings.Contains(o.Status, "Pull complete") {
		return fmt.Sprintf("%s\n", s)
	}

	return fmt.Sprintf("%s\r", s)
}

func handleBuildOutput(b []byte) string {
	logging.GetLogger().Debug("output", zap.ByteString("msg", b))
	type buildOutput struct {
		Stream string `json:"stream"`
	}
	var o buildOutput
	json.Unmarshal(b, &o)
	return o.Stream
}
