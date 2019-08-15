package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/velocity-ci/velocity/backend/pkg/vcli"

	"github.com/spf13/cobra"
	"github.com/velocity-ci/velocity/backend/pkg/git"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/build"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
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
		root, err := config.GetRootConfig()
		if err != nil {
			return err
		}
		pipelines, err := config.GetPipelinesFromRoot(root)
		if err != nil {
			return err
		}
		blueprints, err := config.GetBlueprintsFromRoot(root)
		if err != nil {
			return err
		}

		branch, err := cmd.Flags().GetString("branch")
		if err != nil {
			return err
		}
		if branch == "" {
			branch, err = git.CurrentBranch(root.Path)
			if err != nil {
				return err
			}
		}

		constructionPlan, err := build.NewConstructionPlanFromPipeline(
			args[0],
			pipelines,
			blueprints,
			&vcli.ParameterResolver{},
			nil,
			branch,
			"",
			root.Path,
		)

		switch {
		case runPlanOnly && machineReadable:
			return runConstructionPlanPlanOnlyAndMachineReadable(constructionPlan)
		case runPlanOnly:
			return runConstructionPlanPlanOnly(constructionPlan)
		default:
			return runConstructionPlanText(constructionPlan)
		}
	},
}

func runConstructionPlanPlanOnly(plan *build.ConstructionPlan) error {
	printHeader(plan.Name)
	for _, stage := range plan.Stages {
		fmt.Fprintf(os.Stdout,
			output.ColorFmt(aurora.CyanFg, fmt.Sprintf(" Stage %d", stage.Index), "\n"))
		for _, task := range stage.Tasks {
			fmt.Fprintf(os.Stdout, "   -> %s: %s\n",
				task.Blueprint.Name,
				aurora.Colorize(task.Blueprint.Description, aurora.ItalicFm|aurora.Gray(20, "").Color()),
			)
			for i, step := range task.Steps {
				fmt.Fprintf(os.Stdout, "        %d: %s \n           %s\n",
					i+1,
					step.GetType(),
					aurora.Colorize(strings.ReplaceAll(step.GetDetails(), "\n", "\n           "), aurora.ItalicFm|aurora.Gray(20, "").Color()),
				)
			}
		}
	}
	return nil
}

func printHeader(header string) {
	header = fmt.Sprintf("~ %s ~", header)
	border := ""
	for i := 0; i < len(header); i++ {
		border += "~"
	}

	fmt.Fprintf(os.Stdout, output.ColorFmt(aurora.MagentaFg, border, "\n"))
	fmt.Fprintf(os.Stdout, output.ColorFmt(aurora.MagentaFg, header, "\n"))
	fmt.Fprintf(os.Stdout, output.ColorFmt(aurora.MagentaFg, border, "\n"))
}
