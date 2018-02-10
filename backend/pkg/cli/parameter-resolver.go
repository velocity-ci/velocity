package cli

import (
	"bufio"
	"fmt"
	"os"
)

type ParameterResolver struct {
}

func (pR ParameterResolver) Resolve(paramName string) (string, error) {

	var text string
	reader := bufio.NewReader(os.Stdin)

	fromEnv := os.Getenv(paramName)

	for text == "" {
		if len(fromEnv) > 0 {
			fmt.Printf("Enter value for %s (%s): ", paramName, fromEnv)
		} else {
			fmt.Printf("Enter value for %s: ", paramName)
		}
		text, _ = reader.ReadString('\n')
		if text == "\n" {
			text = fromEnv
		}
	}

	return text, nil
}
