package main

import (
	"github.com/spf13/cobra"
)

func init() {
	runCmd.AddCommand(runPipelineCmd)
}

var runPipelineCmd = &cobra.Command{
	Use:     "pipeline",
	Aliases: []string{"p"},
	Short:   "runs a given pipeline",
	Long:    `runs a given pipeline`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
