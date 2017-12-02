package branch

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type Repository interface {
	Save(b Branch) Branch
	Delete(b Branch)
	GetByProjectIDAndName(projectID string, name string) (Branch, error)
	GetAllByProjectID(projectID string, q Query) ([]Branch, uint64)
}

type Branch struct {
	ID          string    `json:"id" gorm:"primary_key"`
	Name        string    `json:"name"`
	ProjectID   string    `json:"projectId"`
	LastUpdated time.Time `json:"lastUpdated"`
}

func NewBranch(projectID string, name string) Branch {
	return Branch{
		ID:        uuid.NewV3(uuid.NewV1(), projectID).String(),
		ProjectID: projectID,
		Name:      name,
	}
}

type Query struct {
	Amount uint64
	Page   uint64
}

type ManyResponse struct {
	Total  uint64           `json:"total"`
	Result []ResponseBranch `json:"result"`
}

type ResponseBranch struct {
	Name string `json:"name"`
}

func NewResponseBranch(b Branch) ResponseBranch {
	return ResponseBranch{
		Name: b.Name,
	}
}
