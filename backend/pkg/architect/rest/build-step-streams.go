package rest

import "github.com/velocity-ci/velocity/backend/pkg/domain/build"

type streamResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func newStreamResponse(s *build.Stream) *streamResponse {
	return &streamResponse{
		ID:   s.ID,
		Name: s.Name,
	}
}
