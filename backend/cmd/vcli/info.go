package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
)

func init() {
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:     "info",
	Aliases: []string{"i"},
	Short:   "Displays information about the current project",
	Long:    `information about the current project`,
	Args:    cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := config.GetRootConfig()
		if err != nil {
			return err
		}

		switch {
		case machineReadable:
			return infoMachine(root)
		default:
			return infoText(root)
		}

	},
}

func infoMachine(root *config.Root) error {
	jsonBytes, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%s\n", jsonBytes)
	return nil
}

func infoText(root *config.Root) error {
	fmt.Fprintf(os.Stdout, "\n~~ %s ~~\n", output.Italic("Parameters"))
	if len(root.Parameters) > 0 {
		// for _, parameter := range root.Parameters {
		// 	fmt.Fprintf(os.Stdout, "  %s\t%s\n", parameter.Type)
		// }
	} else {
		fmt.Fprintln(os.Stdout, "  none found")
	}

	fmt.Fprintf(os.Stdout, "\n~~ %s ~~\n", output.Italic("Plugins"))
	if len(root.Plugins) > 0 {
		for _, plugin := range root.Plugins {
			fmt.Fprintf(os.Stdout, "  %s\t%s\n", plugin.Use, plugin.Events)
		}
	} else {
		fmt.Fprintln(os.Stdout, "  none found")
	}

	fmt.Fprintln(os.Stdout, "")

	return nil
}
