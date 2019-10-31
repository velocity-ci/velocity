package cmds

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"github.com/velocity-ci/velocity/backend/pkg/grpc/builder"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
)

var grpcOpts struct {
	address  string
	insecure bool
}

func init() {
	rootCmd.AddCommand(grpcCmd)
	grpcCmd.PersistentFlags().StringVar(&grpcOpts.address, "address", "", "the grpc address to connect to")
	grpcCmd.PersistentFlags().BoolVar(&grpcOpts.insecure, "insecure", false, "whether or not to use an insecure grpc connection")
}

var grpcCmd = &cobra.Command{
	Use:     "grpc",
	Aliases: []string{"i"},
	Short:   "Runs GRPC requests",
	Long:    ``,
	// Args:    cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		builder.Insecure = grpcOpts.insecure
		conn, err := builder.NewConn(grpcOpts.address)
		if err != nil {
			return err
		}
		defer conn.Close()

		client := v1.NewProjectServiceClient(conn)

		project, err := client.CreateProject(context.Background(), &v1.CreateProjectRequest{
			Name: "test",
			Repository: &v1.Repository{
				Address: "https://github.com/velocity-ci/velocity.git",
			},
		})

		log.Printf("%+v", project)

		return err
	},
}
