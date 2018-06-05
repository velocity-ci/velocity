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

	step, err := m.stepManager.GetByID(sL.StepID)
	if err != nil {
		glog.Error(err)
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
		glog.Error(err)
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
