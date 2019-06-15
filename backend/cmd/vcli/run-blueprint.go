package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/velocity-ci/velocity/backend/pkg/git"
	"github.com/velocity-ci/velocity/backend/pkg/vcli"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/build"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
)

func init() {
	runCmd.AddCommand(runBlueprintCmd)
}

var runBlueprintCmd = &cobra.Command{
	Use:     "blueprint",
	Aliases: []string{"b"},
	Short:   "runs a given blueprint",
	Long:    `runs a given blueprint`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := config.GetRootConfig()
		if err != nil {
			return err
		}
		blueprints, err := config.GetBlueprintsFromRoot(root)
		if err != nil {
			return err
		}

		branch, err := git.CurrentBranch(root.Path)
		if err != nil {
			return err
		}

		constructionPlan, err := build.NewConstructionPlan(
			args[0],
			blueprints,
			&vcli.ParameterResolver{},
			nil,
			branch,
			"",
			root.Path,
		)
		if err != nil {
			return err
		}

		switch {
		case runPlanOnly && machineReadable:
			return runBlueprintPlanOnlyAndMachineReadable(constructionPlan)
		case runPlanOnly:
			return nil
		default:
			return runBlueprintText(constructionPlan)
		}
	},
}

func runBlueprintText(plan *build.ConstructionPlan) error {
	emitter := vcli.NewEmitter()
	err := plan.Execute(emitter)
	if err != nil {
		return err
	}
	return nil
}

func runBlueprintPlanOnlyAndMachineReadable(plan *build.ConstructionPlan) error {
	jsonBytes, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "%s\n", jsonBytes)
	return nil
}
