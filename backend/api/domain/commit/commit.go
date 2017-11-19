package commit

import (
	"time"

	"github.com/velocity-ci/velocity/backend/api/domain/branch"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
)

type Repository interface {
	SaveToProject(p *project.Project, c *Commit) *Commit
	DeleteFromProject(p *project.Project, c *Commit)
	GetByProjectAndHash(p *project.Project, hash string) (*Commit, error)
	GetAllByProject(p *project.Project, q Query) ([]*Commit, uint64)
}

type Query struct {
	Amount uint64
	Page   uint64
	Branch string
	Author string
}

type Commit struct {
	Hash      string
	Author    string
	CreatedAt time.Time
	Message   string
	Branches  []branch.Branch
}

func NewCommit(
	hash string,
	message string,
	author string,
	date time.Time,
	b branch.Branch,
) *Commit {
	return &Commit{
		Hash:      hash,
		Message:   message,
		Author:    author,
		CreatedAt: date,
		Branches:  []branch.Branch{b},
	}
}

type ResponseCommit struct {
	Hash      string    `json:"hash"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
	Message   string    `json:"message"`
	Branches  []string  `json:"branches"`
}

func NewResponseCommit(c *Commit) *ResponseCommit {
	branches := []string{}
	for _, b := range c.Branches {
		branches = append(branches, b.Name)
	}

	return &ResponseCommit{
		Hash:      c.Hash,
		Author:    c.Author,
		CreatedAt: c.CreatedAt,
		Message:   c.Message,
		Branches:  branches,
	}
}

type ManyResponse struct {
	Total  uint64            `json:"total"`
	Result []*ResponseCommit `json:"result"`
}
