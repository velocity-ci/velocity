package cmds

import (
	"fmt"
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

		err = rootCmd.GenZshCompletionFile(filepath.Join(homeDir, "/.velocityci/completion.zsh"))
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Bash completion script: source %s\n", filepath.Join(homeDir, "/.velocityci/completion.bash"))
		fmt.Fprintf(os.Stdout, "ZSH completion script: source %s\n", filepath.Join(homeDir, "/.velocityci/completion.zsh"))
		return nil
	},
}
