package rest

import (
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type stepResponse struct {
	ID          string            `json:"id"`
	Number      int               `json:"number"`
	VStep       *velocity.Step    `json:"step"`
	Streams     []*streamResponse `json:"streams"`
	Status      string            `json:"status"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	StartedAt   time.Time         `json:"startedAt"`
	CompletedAt time.Time         `json:"completedAt"`
}

func newStepResponse(s *build.Step) *stepResponse {
	streams := []*streamResponse{}
	for _, s := range s.Streams {
		streams = append(streams, newStreamResponse(s))
	}
	return &stepResponse{
		ID:          s.ID,
		Number:      s.Number,
		VStep:       s.VStep,
		Status:      s.Status,
		UpdatedAt:   s.UpdatedAt,
		StartedAt:   s.StartedAt,
		CompletedAt: s.CompletedAt,
	}
}
