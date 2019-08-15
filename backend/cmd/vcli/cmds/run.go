package cmds

import (
	"github.com/spf13/cobra"
)

var (
	runPlanOnly bool
	runBranch   string
)

func init() {
	runCmd.PersistentFlags().BoolVar(&runPlanOnly, "plan-only", false, "Only output the build plan")
	runCmd.PersistentFlags().StringVar(&runBranch, "branch", "", "The branch to run with")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:       "run",
	Aliases:   []string{"r"},
	Short:     "Runs blueprints and pipelines",
	Long:      ``,
	ValidArgs: []string{"blueprint", "pipeline", "b", "p"},
	Args:      cobra.ExactValidArgs(1),
	Run:       func(cmd *cobra.Command, args []string) {},
}
