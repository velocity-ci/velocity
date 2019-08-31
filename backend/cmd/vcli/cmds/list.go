package cmds

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/logrusorgru/aurora"

	"github.com/spf13/cobra"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:       "list",
	Aliases:   []string{"l"},
	Short:     "Lists blueprints and pipelines",
	Long:      `Lists all of blueprints and pipelines`,
	ValidArgs: []string{"blueprints", "pipelines", "b", "p"},
	Args:      cobra.RangeArgs(0, 1),
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

	printHeader("Blueprints")
	if len(blueprints) > 0 {
		for _, blueprint := range blueprints {
			fmt.Fprintf(tabWriter, " %s %s\t%s\n",
				output.ColorFmt(aurora.CyanFg, "->", " "),
				blueprint.Name,
				aurora.Colorize(blueprint.Description, aurora.ItalicFm|aurora.Gray(20, "").Color()),
			)
		}
		tabWriter.Flush()
	} else {
		fmt.Fprintln(os.Stdout, "  none found")
	}

	printHeader("Pipelines")
	if len(pipelines) > 0 {
		for _, pipeline := range pipelines {
			fmt.Fprintf(tabWriter, " %s %s\t%s\n",
				output.ColorFmt(aurora.CyanFg, "->", " "),
				pipeline.Name,
				aurora.Colorize(pipeline.Description, aurora.ItalicFm|aurora.Gray(20, "").Color()),
			)
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
