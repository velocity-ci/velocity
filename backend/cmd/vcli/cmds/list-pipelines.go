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
	listCmd.AddCommand(listPipelinesCmd)
}

var listPipelinesCmd = &cobra.Command{
	Use:     "pipelines",
	Aliases: []string{"p"},
	Short:   "lists pipelines",
	Long:    `lists pipelines`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := config.GetRootConfig()
		if err != nil {
			return err
		}

		pipelines, err := config.GetPipelinesFromRoot(root)
		if err != nil {
			return err
		}

		switch {
		case machineReadable:
			return listPipelinesMachine(pipelines)
		default:
			return listPipelinesText(pipelines)
		}
	},
}

func listPipelinesText(pipelines []*config.Pipeline) error {
	tabWriter := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

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
	return nil
}

func listPipelinesMachine(pipelines []*config.Pipeline) error {
	jsonBytes, err := json.MarshalIndent(pipelines, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%s\n", jsonBytes)
	return nil
}
