package rest

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
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

type stepList struct {
	Total int             `json:"total"`
	Data  []*stepResponse `json:"data"`
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

type buildStepHandler struct {
	buildManager     *build.BuildManager
	buildStepManager *build.StepManager
}

func newBuildStepHandler(
	buildManager *build.BuildManager,
	buildStepManager *build.StepManager,
) *buildStepHandler {
	return &buildStepHandler{
		buildManager:     buildManager,
		buildStepManager: buildStepManager,
	}
}

func (h *buildStepHandler) getStepsForBuildUUID(c echo.Context) error {

	b := getBuildByID(c, h.buildManager)
	if b == nil {
		return nil
	}

	r := []*stepResponse{}
	for _, s := range b.Steps {
		r = append(r, newStepResponse(s))
	}

	c.JSON(http.StatusOK, stepList{
		Total: len(r),
		Data:  r,
	})

	return nil
}
