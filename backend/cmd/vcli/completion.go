package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion scripts",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		os.MkdirAll(filepath.Join(homeDir, "/.velocityci"), 0700)
		err = rootCmd.GenBashCompletionFile(filepath.Join(homeDir, "/.velocityci/completion.bash"))
		if err != nil {
			return err
		}

		return rootCmd.GenZshCompletionFile(filepath.Join(homeDir, "/.velocityci/completion.zsh"))
	},
}
