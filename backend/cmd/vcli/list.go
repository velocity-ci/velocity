package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l"},
	Short:   "Lists blueprints and pipelines",
	Long:    `Lists all of blueprints and pipelines`,
	Args:    cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := config.GetRootConfig()
		if err != nil {
			return err
		}

		blueprints, err := config.GetBlueprintsFromRoot(root)
		if err != nil {
			return err
		}

		pipelines, err := config.GetPipelinesFromRoot(root)
		if err != nil {
			return err
		}

		switch {
		case machineReadable:
			return listMachine(blueprints, pipelines)
		default:
			return listText(blueprints, pipelines)
		}
	},
}

func listText(blueprints []*config.Blueprint, pipelines []*config.Pipeline) error {
	tabWriter := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	italic := color.New(color.Italic).SprintFunc()
	fmt.Fprintf(os.Stdout, "\n~~ %s ~~\n", italic("Blueprints"))
	if len(blueprints) > 0 {
		for _, blueprint := range blueprints {
			fmt.Fprintf(tabWriter, "  %s\t%s\n", blueprint.Name, blueprint.Description)
		}
		tabWriter.Flush()
	} else {
		fmt.Fprintln(os.Stdout, "  none found")
	}

	fmt.Fprintf(os.Stdout, "\n~~ %s ~~\n", italic("Pipelines"))
	if len(pipelines) > 0 {
		for _, pipeline := range pipelines {
			fmt.Fprintf(tabWriter, "  %s\t%s\n", pipeline.Name, pipeline.Description)
		}
		tabWriter.Flush()
	} else {
		fmt.Fprintln(os.Stdout, "  none found")
	}

	fmt.Fprintln(os.Stdout, "")

	return nil
}

func listMachine(blueprints []*config.Blueprint, pipelines []*config.Pipeline) error {
	jsonBytes, err := json.MarshalIndent(
		map[string]interface{}{
			"blueprints": blueprints,
			"pipelines":  pipelines,
		}, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%s\n", jsonBytes)
	return nil
}
