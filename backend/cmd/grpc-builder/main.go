package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/velocity-ci/velocity/backend/pkg/grpc/builder"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/output"
)

// BuildVersion represents the current build tag of this Builder. It is set at compile-time with ldflags
var BuildVersion = "dev"

var (
	noColor  bool
	insecure bool
)

var (
	gracefulStop = make(chan os.Signal)
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", false, "Use an insecure connection to GRPC server")
}

var rootCmd = &cobra.Command{
	Use:   "vbuilder",
	Short: "Velocity Builder",
	Long:  `Builds Velocity tasks`,
	Args:  cobra.ExactArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			output.ColorDisable()
		}
		builder.Insecure = insecure
		signal.Notify(gracefulStop, syscall.SIGTERM)
		signal.Notify(gracefulStop, syscall.SIGINT)
		go func() {
			sig := <-gracefulStop
			fmt.Printf("\ncaught signal: %+v\n", sig)
			close(gracefulStop)
		}()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		address := args[0]
		token, err := builder.Register(address)
		if err != nil {
			return err
		}

		return builder.BreakRoom(address, token, gracefulStop)
	},
	Version: BuildVersion,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stdout, output.ColorFmt(output.ANSIError, "%s", "\n"), err)
		os.Exit(1)
	}
}
