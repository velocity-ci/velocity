package cmds

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
)

func init() {
	listCmd.AddCommand(listBlueprintsCmd)
}

var listBlueprintsCmd = &cobra.Command{
	Use:     "blueprints",
	Aliases: []string{"b"},
	Short:   "lists blueprints",
	Long:    `lists blueprints`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := config.GetRootConfig()
		if err != nil {
			return err
		}

		blueprints, err := config.GetBlueprintsFromRoot(root)
		if err != nil {
			return err
		}

		switch {
		case machineReadable:
			return listBlueprintsMachine(blueprints)
		default:
			return listBlueprintsText(blueprints)
		}
	},
}

func listBlueprintsText(blueprints []*v1.Blueprint) error {
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
	return nil
}

func listBlueprintsMachine(blueprints []*v1.Blueprint) error {
	jsonBytes, err := json.MarshalIndent(blueprints, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%s\n", jsonBytes)
	return nil
}
