package rest

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
)

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

type buildStreamHandler struct {
	buildStepManager   *build.StepManager
	buildStreamManager *build.StreamManager
}

func newBuildStreamHandler(
	buildStepManager *build.StepManager,
	buildStreamManager *build.StreamManager,
) *buildStreamHandler {
	return &buildStreamHandler{
		buildStepManager:   buildStepManager,
		buildStreamManager: buildStreamManager,
	}
}

func (h *buildStreamHandler) getByStepID(c echo.Context) error {
	step := getStepByID(c, h.buildStepManager)
	if step == nil {
		return nil
	}

	r := []*streamResponse{}
	for _, s := range step.Streams {
		r = append(r, newStreamResponse(s))
	}
	return nil
}

func (h *buildStreamHandler) getByID(c echo.Context) error {
	stream := getStreamByID(c, h.buildStreamManager)
	if stream == nil {
		return nil
	}

	return nil
}

func getStreamByID(c echo.Context, buildStreamManager *build.StreamManager) *build.Stream {
	id := c.Param("id")
	s, err := buildStreamManager.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, "not found")
		return nil
	}

	// TODO: replace with stream lines (using file manager)
	c.JSON(http.StatusOK, newStreamResponse(s))
	return nil
}
