package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ParameterResolver struct {
}

func (pR ParameterResolver) Resolve(paramName string) (string, error) {

	var text string
	reader := bufio.NewReader(os.Stdin)

	fromEnv := os.Getenv(paramName)

	for text == "" {
		if len(fromEnv) > 0 {
			fmt.Printf("\nEnter value for %s (%s): ", paramName, fromEnv)
		} else {
			fmt.Printf("\nEnter value for %s: ", paramName)
		}
		text, _ = reader.ReadString('\n')
		if text == "\n" {
			text = fromEnv
		}
	}

	return strings.TrimSpace(text), nil
}
