package commit

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/api/domain/branch"
)

type Repository interface {
	Save(c Commit) Commit
	Delete(c Commit)
	GetByProjectIDAndCommitID(projectID string, commitID string) (Commit, error)
	GetAllByProjectID(projectID string, q Query) ([]Commit, uint64)
}

type Query struct {
	Amount uint64
	Page   uint64
	Branch string
	Author string
}

type Commit struct {
	ID        string          `json:"id" gorm:"primary_key"`
	ProjectID string          `json:"projectId"`
	Hash      string          `json:"hash"`
	Author    string          `json:"author"`
	CreatedAt time.Time       `json:"createdAt"`
	Message   string          `json:"message"`
	Branches  []branch.Branch `json:"branches" gorm:"many2many:commit_branches;AssociationForeignKey:ID;ForeignKey:ID"`
}

func NewCommit(
	projectID string,
	hash string,
	message string,
	author string,
	date time.Time,
	branches []branch.Branch,
) Commit {
	return Commit{
		ID:        uuid.NewV3(uuid.NewV1(), hash).String(),
		ProjectID: projectID,
		Hash:      hash,
		Message:   message,
		Author:    author,
		CreatedAt: date,
		Branches:  branches,
	}
}

type ResponseCommit struct {
	ID        string    `json:"id"`
	Hash      string    `json:"hash"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
	Message   string    `json:"message"`
	Branches  []string  `json:"branches"`
}

func NewResponseCommit(c Commit) ResponseCommit {
	branches := []string{}
	for _, b := range c.Branches {
		branches = append(branches, b.Name)
	}

	return ResponseCommit{
		ID:        c.ID,
		Hash:      c.Hash,
		Author:    c.Author,
		CreatedAt: c.CreatedAt,
		Message:   c.Message,
		Branches:  branches,
	}
}

type ManyResponse struct {
	Total  uint64           `json:"total"`
	Result []ResponseCommit `json:"result"`
}
