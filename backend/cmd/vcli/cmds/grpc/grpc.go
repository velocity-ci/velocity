package grpc

import (
	"github.com/spf13/cobra"
)

var grpcOpts struct {
	address  string
	insecure bool
}

func init() {
	Cmd.PersistentFlags().StringVar(&grpcOpts.address, "address", "", "the grpc address to connect to")
	Cmd.PersistentFlags().BoolVar(&grpcOpts.insecure, "insecure", false, "whether or not to use an insecure grpc connection")
	Cmd.MarkPersistentFlagRequired("address")
}

var Cmd = &cobra.Command{
	Use:   "grpc",
	Short: "Runs GRPC requests",
	Long:  ``,
	// Args:    cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
