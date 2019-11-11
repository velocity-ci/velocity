package grpc

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"github.com/velocity-ci/velocity/backend/pkg/grpc/builder"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
)

var createProjectOpts struct {
	name              string
	repositoryAddress string
}

func init() {
	projectCreateCmd.Flags().StringVarP(&createProjectOpts.repositoryAddress, "repository-address", "a", "", "The repository address")
	projectCreateCmd.Flags().StringVarP(&createProjectOpts.name, "name", "n", "", "The project name")
	projectCreateCmd.MarkFlagRequired("repository-address")
	projectCreateCmd.MarkFlagRequired("name")
	projectCmd.AddCommand(projectCreateCmd)
}

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new Project via GRPC",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		builder.Insecure = grpcOpts.insecure
		conn, err := builder.NewConn(grpcOpts.address)
		if err != nil {
			return err
		}
		defer conn.Close()

		client := v1.NewProjectServiceClient(conn)

		project, err := client.CreateProject(context.Background(), &v1.CreateProjectRequest{
			Name: createProjectOpts.name,
			Repository: &v1.Repository{
				Address: createProjectOpts.repositoryAddress,
			},
		})

		log.Printf("%+v", project)

		return err
	},
}
