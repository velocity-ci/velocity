package project

import (
	"encoding/json"

	"github.com/gosimple/slug"
)

type Project struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Repository GitRepository `json:"repository"`
}

func (p *Project) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

func NewProject(name string, repository GitRepository) Project {
	return Project{
		ID:         slug.Make(name),
		Name:       name,
		Repository: repository,
	}
}

type GitRepository struct {
	Address    string `json:"address"`
	PrivateKey string `json:"privateKey"`
}
