package secretary

import (
	"fmt"

	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

// get commits
/*
 {
	 "repository": {},
	 "query": {
		 "limit": 10,
		 "page": 1,

	 }
 }

 {
	 "topic": "secretaries:pool",
	 "event": "phx_reply",
	 "ref": 32,
	 "payload": {
		 "commits": [
			 {
				"sha": "",
				"author": "",
				"created_at": "",
				"message": ""
			 }
		 ]
	 }

 }
*/

type Repository struct {
	Address        string `json:"address"`
	PrivateKey     string `json:"privateKey"`
	KnownHostEntry string `json:"knownHostEntry"`
}

func getRepository(r *Repository) (*velocity.RawRepository, error) {
	rawRepo, err := velocity.Clone(
		&velocity.GitRepository{
			Address:    r.Address,
			PrivateKey: r.PrivateKey,
		},
		&velocity.BlankWriter{},
		&velocity.CloneOptions{
			Bare:      true,
			Full:      true,
			Submodule: false,
		},
	)

	if err != nil {
		return nil, err
	}

	return rawRepo, err
}

type GetCommitsQuery struct {
	Limit    uint64   `json:"limit"`
	Page     uint64   `json:"page"`
	Branch   string   `json:"branch"`
	Branches []string `json:"branches"`
}

type GetCommitsInput struct {
	Repository *Repository      `json:"repository"`
	Query      *GetCommitsQuery `json:"query"`
}

type Commit struct {
	SHA string `json:"sha"`
}

type GetCommitsOutput struct {
	Commits []*Commit `json:"commits"`
	Total   uint64    `json:"total"`
}

func getCommitsEvent(payload interface{}) (interface{}, error) {
	if input, ok := payload.(GetCommitsInput); ok {
		return getCommits(&input)
	}
	return nil, fmt.Errorf("invalid payload")
}

func getCommits(input *GetCommitsInput) (*GetCommitsOutput, error) {
	rawRepo, err := getRepository(input.Repository)
	if err != nil {
		return nil, err
	}

	out := &GetCommitsOutput{}
	out.Total = rawRepo.GetTotalCommits()

	return nil, nil
}
