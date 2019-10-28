package architect

import (
	"context"
	"fmt"

	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
)

type ProjectServer struct{}

func NewProjectServer() *ProjectServer {
	return &ProjectServer{}
}

func (s *ProjectServer) CreateProject(ctx context.Context, req *v1.CreateProjectRequest) (*v1.Project, error) {

	return nil, fmt.Errorf("TODO")
}

func (s *ProjectServer) GetProject(ctx context.Context, req *v1.GetProjectRequest) (*v1.Project, error) {

	return nil, fmt.Errorf("TODO")
}

func (s *ProjectServer) ListProjects(ctx context.Context, req *v1.ListProjectsRequest) (*v1.ListProjectsResponse, error) {

	return nil, fmt.Errorf("TODO")
}
