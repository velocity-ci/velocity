package architect

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/git"
	"github.com/velocity-ci/velocity/backend/pkg/grpc/architect/db"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
)

type ProjectServer struct {
	db                *db.DB
	repositoryManager *git.RepositoryManager
}

func NewProjectServer(
	db *db.DB,
	repositoryManager *git.RepositoryManager,
) *ProjectServer {
	return &ProjectServer{
		db:                db,
		repositoryManager: repositoryManager,
	}
}

func (s *ProjectServer) CreateProject(ctx context.Context, req *v1.CreateProjectRequest) (*v1.Project, error) {
	p := &v1.Project{
		Id:         uuid.NewV4().String(),
		Name:       req.GetName(),
		Repository: req.GetRepository(),
		CreatedAt:  ptypes.TimestampNow(),
		UpdatedAt:  ptypes.TimestampNow(),
	}

	p, err := s.db.CreateProject(ctx, p)
	if err != nil {
		return nil, err
	}

	go s.repositoryManager.Add(p.GetRepository().GetAddress(), p.GetRepository().GetSshConfig().GetPrivateKey(), p.GetRepository().GetSshConfig().GetHostKey())

	return p, nil
}

func (s *ProjectServer) GetProject(ctx context.Context, req *v1.GetProjectRequest) (*v1.Project, error) {
	p, err := s.db.GetProjectByUUID(ctx, req.GetProjectId())
	return p, err
}

func (s *ProjectServer) ListProjects(ctx context.Context, req *v1.ListProjectsRequest) (*v1.ListProjectsResponse, error) {
	ps, err := s.db.GetProjects(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListProjectsResponse{
		Projects: ps,
	}, nil
}

func (s *ProjectServer) GetHead(context.Context, *v1.GetHeadRequest) (*v1.Head, error) {
	return nil, fmt.Errorf("TODO")
}
func (s *ProjectServer) GetCommit(context.Context, *v1.GetCommitRequest) (*v1.Commit, error) {
	return nil, fmt.Errorf("TODO")
}
func (s *ProjectServer) ListHeads(context.Context, *v1.ListHeadsRequest) (*v1.ListHeadsResponse, error) {
	return nil, fmt.Errorf("TODO")
}
func (s *ProjectServer) ListCommits(context.Context, *v1.ListCommitsRequest) (*v1.ListCommitsResponse, error) {
	return nil, fmt.Errorf("TODO")
}
