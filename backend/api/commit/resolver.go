package commit

import (
	"net/http"
	"strconv"
)

func QueryOptsFromRequest(r *http.Request) *CommitQueryOpts {
	reqQueries := r.URL.Query()

	amount := 15
	if a, err := strconv.Atoi(reqQueries.Get("amount")); err == nil {
		amount = a
	}
	page := 1

	if p, err := strconv.Atoi(reqQueries.Get("page")); err == nil {
		page = p
	}

	return &CommitQueryOpts{
		Branch: reqQueries.Get("branch"),
		Amount: amount,
		Page:   page,
	}
}
