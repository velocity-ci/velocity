package secretary

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/gosimple/slug"
	"github.com/velocity-ci/velocity/backend/pkg/git"
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

var Workspace = "/opt/velocityci/repositories"

func init() {
	u, _ := user.Current()
	if u.Uid != "0" {
		wd, _ := os.Getwd()
		Workspace = fmt.Sprintf("%s/_velocity_data/repositories", wd)
	}
}

type Repository struct {
	Address        string `json:"address"`
	PrivateKey     string `json:"privateKey"`
	KnownHostEntry string `json:"knownHostEntry"`
}

func getRepository(r *Repository) (*git.RawRepository, error) {
	repoDir := filepath.Join(Workspace, slug.Make(r.Address))

	rawRepo, err := git.Clone(
		&git.Repository{
			Address:    r.Address,
			PrivateKey: r.PrivateKey,
		},
		&git.CloneOptions{
			// Bare:      true,
			Depth:     0,
			Submodule: true,
		},
		repoDir,
		nil,
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

func getCommitsEvent(payload json.RawMessage) (interface{}, error) {
	input := GetCommitsInput{}
	json.Unmarshal(payload, &input)
	return getCommits(&input)
	// return nil, fmt.Errorf("invalid payload: %s", payload)
}

func getCommits(input *GetCommitsInput) (*GetCommitsOutput, error) {
	rawRepo, err := getRepository(input.Repository)
	if err != nil {
		return nil, err
	}

	out := &GetCommitsOutput{}
	out.Total = rawRepo.GetTotalCommits()

	return out, nil
}
