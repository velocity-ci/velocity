package domain

import (
	"time"
)

type Project struct {
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	ID         string    `json:"id" gorm:"primary_key"`
	Name       string    `json:"name"`
	Repository string    `json:"repository"`
	PrivateKey string    `json:"key"`
}
