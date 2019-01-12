package secretary

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

type GetCommitsQuery struct {
	Limit    uint64   `json:"limit"`
	Page     uint64   `json:"page"`
	Branch   string   `json:"branch"`
	Branches []string `json:"branches"`
}

type GetCommitsInput struct {
	Repository string           `json:"repository"`
	Query      *GetCommitsQuery `json:"query"`
}

type Commit struct {
	SHA string `json:"sha"`
}

type GetCommitsOutput struct {
	Commits []*Commit `json:"commits"`
	Total   uint64    `json:"total"`
}

func getCommits(input *GetCommitsInput) (res *GetCommitsOutput, err error) {

	return nil, nil
}
