package out

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
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
}

func handleLogOutput(b []byte) string {
	if len(b) <= 8 {
		return ""
	}
	return string(b[8:])
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
		strings.Contains(o.Status, "Pull complete") {
		return s
	}

	return fmt.Sprintf("%s\r", s)
}

func handleBuildOutput(b []byte) string {
	type buildOutput struct {
		Stream string `json:"stream"`
	}
	var o buildOutput
	json.Unmarshal(b, &o)
	return strings.TrimSpace(o.Stream)
}
