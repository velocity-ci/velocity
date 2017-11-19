package branch

import "github.com/velocity-ci/velocity/backend/api/domain/project"

type Repository interface {
	SaveToProject(p *project.Project, c *Branch) *Branch
	DeleteFromProject(p *project.Project, c *Branch)
	GetByProjectAndName(p *project.Project, hash string) (*Branch, error)
	GetAllByProject(p *project.Project, q Query) ([]*Branch, uint64)
}

type Branch struct {
	Name string
}

func NewBranch(name string) *Branch {
	return &Branch{
		Name: name,
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
