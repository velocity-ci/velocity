package rest

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
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

type stepList struct {
	Total int             `json:"total"`
	Data  []*stepResponse `json:"data"`
}

func newStepResponse(s *build.Step, streams []*streamResponse) *stepResponse {
	return &stepResponse{
		ID:          s.ID,
		Number:      s.Number,
		VStep:       s.VStep,
		Status:      s.Status,
		UpdatedAt:   s.UpdatedAt,
		StartedAt:   s.StartedAt,
		CompletedAt: s.CompletedAt,
		Streams:     streams,
	}
}

func stepsToStepResponse(steps []*build.Step, streamManager *build.StreamManager) (r []*stepResponse) {
	for _, s := range steps {
		streams := streamManager.GetStreamsForStep(s)
		rStreams := streamsToStreamResponse(streams)
		r = append(r, newStepResponse(s, rStreams))
	}
	return r
}

type buildStepHandler struct {
	buildManager       *build.BuildManager
	buildStepManager   *build.StepManager
	buildStreamManager *build.StreamManager
}

func newBuildStepHandler(
	buildManager *build.BuildManager,
	buildStepManager *build.StepManager,
	buildStreamManager *build.StreamManager,
) *buildStepHandler {
	return &buildStepHandler{
		buildManager:       buildManager,
		buildStepManager:   buildStepManager,
		buildStreamManager: buildStreamManager,
	}
}

func (h *buildStepHandler) getStepsForBuildID(c echo.Context) error {

	b := getBuildByID(c, h.buildManager)
	if b == nil {
		return nil
	}

	steps := h.buildStepManager.GetStepsForBuild(b)

	r := stepsToStepResponse(steps, h.buildStreamManager)

	c.JSON(http.StatusOK, stepList{
		Total: len(r),
		Data:  r,
	})

	return nil
}

func (h *buildStepHandler) getByID(c echo.Context) error {
	s := getStepByID(c, h.buildStepManager)
	if s == nil {
		return nil
	}
	streams := h.buildStreamManager.GetStreamsForStep(s)

	c.JSON(http.StatusOK, newStepResponse(s, streamsToStreamResponse(streams)))
	return nil
}

func getStepByID(c echo.Context, buildStepManager *build.StepManager) *build.Step {
	id := c.Param("id")
	s, err := buildStepManager.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, "not found")
		return nil
	}
	return s
}
