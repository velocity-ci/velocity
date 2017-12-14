package auth

import (
	"time"
)

type UserAuth struct {
	Username string    `json:"username"`
	Token    string    `json:"authToken"`
	Expires  time.Time `json:"expires"`
}
