package builder

import (
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

func (m *Manager) monitor(b *Builder) {
	for {
		message := BuilderRespMessage{}
		err := b.ws.ReadJSON(&message)
		if err != nil {
			velocity.GetLogger().Error("could not read websocket message", zap.Error(err))
			m.Delete(b)
			b.ws.Close()
			return
		}

		switch message.Type {
		case "log":
			m.builderLogMessage(message.Data.(*BuilderStreamLineMessage), b)
			break
		default:
			velocity.GetLogger().Error("got invalid message type from builder", zap.String("message type", message.Type))
		}

	}
}

func (m *Manager) builderLogMessage(sL *BuilderStreamLineMessage, builder *Builder) {
	stream, err := m.streamManager.GetByID(sL.StreamID)
	if err != nil {
		velocity.GetLogger().Error("could not get stream", zap.String("streamID", sL.StreamID), zap.Error(err))
		return
	}

	step, err := m.stepManager.GetByID(sL.StepID)
	if err != nil {
		velocity.GetLogger().Error("could not get step", zap.String("streamID", sL.StepID), zap.Error(err))
		return
	}

	m.streamManager.CreateStreamLine(stream,
		sL.LineNumber,
		time.Now().UTC(),
		sL.Output,
	)

	if step.Status == velocity.StateWaiting {
		step.Status = sL.Status
		step.StartedAt = time.Now().UTC()
		m.stepManager.Update(step)
	}

	if sL.Status == velocity.StateSuccess || sL.Status == velocity.StateFailed {
		step.Status = sL.Status
		step.CompletedAt = time.Now().UTC()
		m.stepManager.Update(step)
	}

	b, err := m.buildManager.GetBuildByID(sL.BuildID)
	if err != nil {
		velocity.GetLogger().Error("could not get build", zap.String("buildID", sL.BuildID), zap.Error(err))
		return
	}
	steps := m.stepManager.GetStepsForBuild(b)

	if b.StartedAt.IsZero() {
		b.Status = sL.Status
		b.StartedAt = time.Now().UTC()
		m.buildManager.Update(b)
	}

	if step.Number == (len(steps)-1) && sL.Status == velocity.StateSuccess || sL.Status == velocity.StateFailed {
		b.Status = sL.Status
		b.CompletedAt = time.Now().UTC()
		m.buildManager.Update(b)

		builder.State = stateReady
		m.Save(builder)
	}

}
