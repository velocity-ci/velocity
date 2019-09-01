package cmds

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
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
	Args:      cobra.OnlyValidArgs,
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

	if err := listBlueprintsText(blueprints); err != nil {
		return err
	}

	if err := listPipelinesText(pipelines); err != nil {
		return err
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
