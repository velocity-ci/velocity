package rest

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
)

type streamResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type streamList struct {
	Total int               `json:"total"`
	Data  []*streamResponse `json:"data"`
}

func newStreamResponse(s *build.Stream) *streamResponse {
	return &streamResponse{
		ID:   s.ID,
		Name: s.Name,
	}
}

type streamLineResponse struct {
	LineNumber int       `json:"lineNumber"`
	Timestamp  time.Time `json:"timestamp"`
	Output     string    `json:"output"`
}

func newStreamLineResponse(s *build.StreamLine) *streamLineResponse {
	return &streamLineResponse{
		LineNumber: s.LineNumber,
		Timestamp:  s.Timestamp,
		Output:     s.Output,
	}
}

type streamLineList struct {
	Total int                   `json:"total"`
	Data  []*streamLineResponse `json:"data"`
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

	c.JSON(http.StatusOK, &streamList{
		Total: len(r),
		Data:  r,
	})
	return nil
}

func (h *buildStreamHandler) getByID(c echo.Context) error {
	stream := getStreamByID(c, h.buildStreamManager)
	if stream == nil {
		return nil
	}

	c.JSON(http.StatusOK, newStreamResponse(stream))
	return nil
}

func (h *buildStreamHandler) getLogByID(c echo.Context) error {
	stream := getStreamByID(c, h.buildStreamManager)
	if stream == nil {
		return nil
	}

	pQ := getPagingQueryParams(c)
	if pQ == nil {
		return nil
	}
	streamLines, total := h.buildStreamManager.GetStreamLines(stream, pQ)

	rSL := []*streamLineResponse{}
	for _, sL := range streamLines {
		rSL = append(rSL, newStreamLineResponse(sL))
	}

	c.JSON(http.StatusOK, &streamLineList{
		Total: total,
		Data:  rSL,
	})
	return nil
}

func getStreamByID(c echo.Context, buildStreamManager *build.StreamManager) *build.Stream {
	id := c.Param("id")
	s, err := buildStreamManager.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, "not found")
		return nil
	}

	return s
}
