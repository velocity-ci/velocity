package builder

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/velocity-ci/velocity/backend/pkg/builder"
	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

func NewMonitor(
	builderManager *Manager,
	streamManager *build.StreamManager,
	stepManager *build.StepManager,
	buildManager *build.BuildManager,
) *Monitor {
	return &Monitor{
		builderManager: builderManager,
		streamManager:  streamManager,
		stepManager:    stepManager,
		buildManager:   buildManager,
	}
}

type Monitor struct {
	builderManager *Manager
	streamManager  *build.StreamManager
	stepManager    *build.StepManager
	buildManager   *build.BuildManager
	builder        *Builder
}

func (m *Monitor) GetCustomEvents() map[string]func(*phoenix.PhoenixMessage) error {
	return map[string]func(*phoenix.PhoenixMessage) error{
		builder.EventNewStreamLines: m.newStreamLines,
	}
}

func (m *Monitor) Authenticate(s *phoenix.Server, token *jwt.Token, topic string) error {
	parts := strings.Split(topic, ":")
	if len(parts) < 2 {
		return fmt.Errorf("invalid topic %s", topic)
	}
	bldr, err := m.builderManager.GetByID(parts[1])
	if err != nil {
		return err
	}
	if token.Claims.(*jwt.StandardClaims).Subject != bldr.ID {
		return fmt.Errorf("mismatched token %s != %s", token.Claims.(*jwt.StandardClaims).Subject, bldr.ID)
	}

	bldr.WS = s
	bldr.State = StateReady
	m.builder = bldr

	m.builderManager.Save(m.builder)
	return nil
}

func (m *Monitor) newStreamLines(mess *phoenix.PhoenixMessage) error {
	buildLogs := builder.BuildLogPayload{}
	err := json.Unmarshal(mess.Payload.(json.RawMessage), &buildLogs)
	if err != nil {
		return err
	}

	// get stream
	stream, err := m.streamManager.GetByID(buildLogs.StreamID)
	if err != nil {
		velocity.GetLogger().Error("could not get stream", zap.String("streamID", buildLogs.StreamID), zap.Error(err))
		return err
	}

	lines := buildLogs.Lines
	lastLine := lines[len(lines)-1]
	// update stream
	if stream.Status != lastLine.Status {
		stream.Status = lastLine.Status
		m.streamManager.Update(stream)
	}

	for _, line := range lines {
		m.streamManager.CreateStreamLine(stream,
			line.LineNumber,
			line.Timestamp,
			line.Output,
		)
	}

	step, err := m.stepManager.GetByID(stream.Step.ID)
	if err != nil {
		velocity.GetLogger().Error("could not get step", zap.String("streamID", stream.Step.ID), zap.Error(err))
		return err
	}

	// update step
	if step.Status == velocity.StateWaiting {
		step.Status = velocity.StateRunning
		step.StartedAt = time.Now().UTC()
		m.stepManager.Update(step)
	}

	if stream.Status == velocity.StateSuccess || stream.Status == velocity.StateFailed {
		stepStreams := m.streamManager.GetStreamsForStep(step)
		status := velocity.StateSuccess
		for _, stream := range stepStreams {
			if stream.Status != velocity.StateSuccess {
				status = stream.Status
				break
			}
		}
		step.Status = status
		if step.Status == velocity.StateSuccess || step.Status == velocity.StateFailed {
			step.CompletedAt = time.Now().UTC()
		}
		m.stepManager.Update(step)
	}

	b, err := m.buildManager.GetBuildByID(step.Build.ID)
	if err != nil {
		velocity.GetLogger().Error("could not get build", zap.String("buildID", step.Build.ID), zap.Error(err))
		return err
	}
	steps := m.stepManager.GetStepsForBuild(b)

	if b.StartedAt.IsZero() {
		b.Status = lastLine.Status
		b.StartedAt = time.Now().UTC()
		m.buildManager.Update(b)
	}

	// if last step and got success/fail check if other streams are success/fail
	if step.Number == (len(steps)-1) && step.Status == velocity.StateSuccess || step.Status == velocity.StateFailed {
		b.Status = step.Status
		b.CompletedAt = time.Now().UTC()
		m.buildManager.Update(b)

		m.builder.State = StateReady
		m.builderManager.Save(m.builder)
	}

	m.builder.WS.Socket.ReplyOK(mess)
	return nil
}
