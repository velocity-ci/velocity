package velocity

type Build struct {
	Project    *Project `json:"project"`
	CommitHash string   `json:"commit"`
	ID         uint64   `json:"id"`
}

func NewBuild(p *Project, c string, ID uint64) *Build {
	return &Build{
		Project:    p,
		CommitHash: c,
		ID:         ID,
	}
}
