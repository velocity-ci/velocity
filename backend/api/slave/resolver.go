package slave

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

func FromRequest(b io.ReadCloser) (*RequestSlave, error) {
	reqSlave := &RequestSlave{}

	err := json.NewDecoder(b).Decode(reqSlave)
	if err != nil {
		return nil, err
	}

	return reqSlave, nil
}

func QueryOptsFromRequest(r *http.Request) SlaveQuery {
	reqQueries := r.URL.Query()

	amount := uint64(15)
	if a, err := strconv.ParseUint(reqQueries.Get("amount"), 10, 64); err == nil {
		amount = a
	}

	page := uint64(1)
	if p, err := strconv.ParseUint(reqQueries.Get("page"), 10, 64); err == nil {
		page = p
	}

	return SlaveQuery{
		Status: reqQueries.Get("status"),
		Amount: amount,
		Page:   page,
	}
}
