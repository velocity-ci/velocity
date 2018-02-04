package rest

import "github.com/velocity-ci/velocity/backend/pkg/domain/build"

type streamResponse struct {
	UUID string `json:"id"`
	Name string `json:"name"`
}

func newStreamResponse(s *build.Stream) *streamResponse {
	return &streamResponse{
		UUID: s.UUID,
		Name: s.Name,
	}
}
