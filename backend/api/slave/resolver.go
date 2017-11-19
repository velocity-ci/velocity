package slave

// import (
// 	"encoding/json"
// 	"io"
// )

// func FromRequest(b io.ReadCloser) (*RequestSlave, error) {
// 	reqSlave := &RequestSlave{}

// 	err := json.NewDecoder(b).Decode(reqSlave)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return reqSlave, nil
// }
