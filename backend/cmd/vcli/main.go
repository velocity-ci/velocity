package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

// BuildVersion represents the current build tag of this CLI. It is set at compile-time with ldflags
var BuildVersion = "dev"

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stdout, color.RedString("%s\n", err))
		os.Exit(1)
	}
}
