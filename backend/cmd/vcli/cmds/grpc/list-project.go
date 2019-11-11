package grpc

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/velocity-ci/velocity/backend/pkg/grpc/builder"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
)

var listProjectOpts struct {
}

func init() {
	projectCmd.AddCommand(projectListCmd)
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Projects via GRPC",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		builder.Insecure = grpcOpts.insecure
		conn, err := builder.NewConn(grpcOpts.address)
		if err != nil {
			return err
		}
		defer conn.Close()

		client := v1.NewProjectServiceClient(conn)

		resp, err := client.ListProjects(context.Background(), &v1.ListProjectsRequest{})

		for _, p := range resp.GetProjects() {
			fmt.Println(p.GetName())
		}

		return err
	},
}
