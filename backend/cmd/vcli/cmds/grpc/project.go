package grpc

import (
	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(projectCmd)
}

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Runs Project related GRPC requests",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
