package builder

import (
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/velocity"

	"github.com/golang/glog"
)

func (m *Manager) monitor(b *Builder) {
	for {
		message := BuilderRespMessage{}
		err := b.ws.ReadJSON(&message)
		if err != nil {
			glog.Error(err)
			glog.Infof("closing builder websocket: %s", b.ID)
			m.Delete(b)
			b.ws.Close()
			return
		}

		switch message.Type {
		case "log":
			m.builderLogMessage(message.Data.(*BuilderStreamLineMessage), b)
			break
		default:
			glog.Errorf("invalid message type from builder: %s", message.Type)
		}

	}
}

func (m *Manager) builderLogMessage(sL *BuilderStreamLineMessage, builder *Builder) {
	stream, err := m.streamManager.GetByID(sL.StreamID)
	if err != nil {
		glog.Error(err)
		return
	}

	if stream.Status != sL.Status {
		stream.Status = sL.Status
		m.streamManager.Update(stream)
	}

	m.streamManager.CreateStreamLine(stream,
		sL.LineNumber,
		time.Now().UTC(),
		sL.Output,
	)

	step, err := m.stepManager.GetByID(sL.StepID)
	if err != nil {
		glog.Error(err)
		return
	}

	if step.Status == velocity.StateWaiting {
		step.Status = velocity.StateRunning
		step.StartedAt = time.Now().UTC()
		m.stepManager.Update(step)
	}

	// update step
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

	b, err := m.buildManager.GetBuildByID(sL.BuildID)
	if err != nil {
		glog.Error(err)
		return
	}
	steps := m.stepManager.GetStepsForBuild(b)

	if b.StartedAt.IsZero() {
		b.Status = sL.Status
		b.StartedAt = time.Now().UTC()
		m.buildManager.Update(b)
	}

	// if last step and got success/fail check if other streams are success/fail

	if step.Number == (len(steps)-1) && step.Status == velocity.StateSuccess || step.Status == velocity.StateFailed {
		b.Status = step.Status
		b.CompletedAt = time.Now().UTC()
		m.buildManager.Update(b)

		builder.State = stateReady
		m.Save(builder)
	}

}
