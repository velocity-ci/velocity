package branch

import (
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
)

type Repository interface {
	Save(b *Branch) *Branch
	Delete(b *Branch)
	GetByProjectAndName(p *project.Project, hash string) (*Branch, error)
	GetAllByProject(p *project.Project, q Query) ([]*Branch, uint64)
}

type Branch struct {
	ID      string
	Name    string
	Project project.Project
}

func NewBranch(p *project.Project, name string) *Branch {
	return &Branch{
		ID:      uuid.NewV3(uuid.NewV1(), p.ID).String(),
		Project: *p,
		Name:    name,
	}
}

type Query struct {
	Amount uint64
	Page   uint64
}

type ManyResponse struct {
	Total  uint64            `json:"total"`
	Result []*ResponseBranch `json:"result"`
}

type ResponseBranch struct {
	Name string `json:"name"`
}

func NewResponseBranch(b *Branch) *ResponseBranch {
	return &ResponseBranch{
		Name: b.Name,
	}
}
