package architect

import (
	"context"
	"fmt"

	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
)

type RepositoryServer struct{}

func NewRepositoryServer() *RepositoryServer {
	return &RepositoryServer{}
}

func (s *RepositoryServer) GetHead(context.Context, *v1.GetHeadRequest) (*v1.Head, error) {
	return nil, fmt.Errorf("TODO")
}
func (s *RepositoryServer) GetCommit(context.Context, *v1.GetCommitRequest) (*v1.Commit, error) {
	return nil, fmt.Errorf("TODO")
}
func (s *RepositoryServer) ListHeads(context.Context, *v1.ListHeadsRequest) (*v1.ListHeadsResponse, error) {
	return nil, fmt.Errorf("TODO")
}
func (s *RepositoryServer) ListCommits(context.Context, *v1.ListCommitsRequest) (*v1.ListCommitsResponse, error) {
	return nil, fmt.Errorf("TODO")
}
