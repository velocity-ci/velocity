package rest

import (
	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"

	"github.com/Sirupsen/logrus"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
)

type broker struct {
	clients map[string]*Client

	branchManager *githistory.BranchManager
}

func NewBroker(branchManager *githistory.BranchManager) *broker {
	return &broker{
		clients:       map[string]*Client{},
		branchManager: branchManager,
	}
}

func (m *broker) save(c *Client) {
	m.clients[c.ID] = c
}

func (m *broker) remove(c *Client) {
	delete(m.clients, c.ID)
}

func (m *broker) EmitAll(message *domain.Emit) {
	// clientCount := 0
	mess := m.handleEmit(message)
	for _, c := range m.clients {
		if !c.alive {
			m.remove(c)
			break
		}
		for _, s := range c.subscriptions {
			if s == message.Topic {
				err := c.ws.WriteJSON(mess)
				// 			clientCount++
				if err != nil {
					logrus.Println(err)
				}
			}
		}
	}
	// log.Printf("Emitted %s to %d clients", message.Topic, clientCount)
}

func (m *broker) handleEmit(em *domain.Emit) *PhoenixMessage {
	var payload interface{}

	switch v := em.Payload.(type) {
	case *user.User:
		break
	case *knownhost.KnownHost:
		payload = newKnownHostResponse(v)
		break
	case *project.Project:
		payload = newProjectResponse(v)
		break
	case *githistory.Branch:
		payload = newBranchResponse(v)
		break
	case *githistory.Commit:
		bs, _ := m.branchManager.GetAllForCommit(v, domain.NewPagingQuery())
		payload = newCommitResponse(v, bs)
		break
	case *task.Task:
		payload = newTaskResponse(v)
		break
	case *build.Build:
		payload = newBuildResponse(v)
		break
	case *build.StreamLine:
		payload = newStreamLineResponse(v)
		break
	default:
		logrus.Errorf("could not resolve websocket payload %+v", v)
	}

	return &PhoenixMessage{
		Topic:   em.Topic,
		Event:   em.Event,
		Payload: payload,
	}
}
