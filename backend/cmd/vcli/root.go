package main

import (
	"github.com/spf13/cobra"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
)

var (
	noColor         bool
	machineReadable bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	rootCmd.PersistentFlags().BoolVar(&machineReadable, "machine-readable", false, "Output in machine readable format (JSON)")
}

var rootCmd = &cobra.Command{
	Use:   "vcli",
	Short: "Velocity CLI",
	Long:  `Runs Velocity CI blueprints and pipelines locally`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			output.ColorDisable()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	Version: BuildVersion,
}
