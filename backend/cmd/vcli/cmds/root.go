package cmds

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/build"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
)

// BuildVersion represents the current build tag of this CLI. It is set at compile-time with ldflags
var BuildVersion = "dev"

var (
	noColor         bool
	machineReadable bool
)

var (
	gracefulStop = make(chan os.Signal)
	action       build.Stoppable
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	rootCmd.PersistentFlags().BoolVar(&machineReadable, "machine-readable", false, "Output in machine readable format (JSON)")
}

func Execute() error {
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "vcli",
	Short: "Velocity CLI",
	Long:  `Runs Velocity CI blueprints and pipelines locally`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			output.ColorDisable()
		}
		signal.Notify(gracefulStop, syscall.SIGTERM)
		signal.Notify(gracefulStop, syscall.SIGINT)
		go func() {
			sig := <-gracefulStop
			fmt.Printf("\ncaught signal: %+v\n", sig)
			// fmt.Println("Wait for 2 second to finish processing")
			// time.Sleep(2 * time.Second)
			action.Stop()
			// os.Exit(0)
		}()
	},
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
	Version: BuildVersion,
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
