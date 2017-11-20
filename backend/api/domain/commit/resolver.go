package commit

import (
	"net/http"
	"strconv"
)

func QueryOptsFromRequest(r *http.Request) Query {
	reqQueries := r.URL.Query()

	amount := uint64(15)
	if a, err := strconv.ParseUint(reqQueries.Get("amount"), 10, 64); err == nil {
		amount = a
	}

	page := uint64(1)
	if p, err := strconv.ParseUint(reqQueries.Get("page"), 10, 64); err == nil {
		page = p
	}

	return Query{
		Branch: reqQueries.Get("branch"),
		Amount: amount,
		Page:   page,
	}
}
