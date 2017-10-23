package commit

import (
	"fmt"
	"strings"
	"time"
)

type Commit struct {
	Branch  string    `json:"branch"`
	Hash    string    `json:"hash"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
	Message string    `json:"message"`
}

func (c *Commit) OrderedID() string {
	return strings.Join([]string{fmt.Sprintf("%v", c.Date.Unix()), c.Hash[:7]}, "-")
}
