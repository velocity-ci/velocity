package cmds

import (
	"github.com/spf13/cobra"
)

func init() {
	listCmd.AddCommand(listBlueprintsCmd)
}

var listBlueprintsCmd = &cobra.Command{
	Use:     "blueprint",
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
			return listMachine(blueprints, pipelines)
		default:
			return listText(blueprints, pipelines)
		}
	},
}
