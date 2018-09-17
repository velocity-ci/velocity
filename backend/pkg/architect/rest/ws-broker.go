package rest

import (
	"fmt"
	"sync"

	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"

	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"

	"github.com/velocity-ci/velocity/backend/pkg/domain"
)

type broker struct {
	clients map[string]*phoenix.Server
	lock    sync.RWMutex

	branchManager *githistory.BranchManager
	stepManager   *build.StepManager
	streamManager *build.StreamManager
}

func NewBroker(
	branchManager *githistory.BranchManager,
	stepManager *build.StepManager,
	streamManager *build.StreamManager,
) *broker {
	return &broker{
		clients:       map[string]*phoenix.Server{},
		lock:          sync.RWMutex{},
		branchManager: branchManager,
		stepManager:   stepManager,
		streamManager: streamManager,
	}
}

func (m *broker) save(c *phoenix.Server) {
	m.lock.Lock()
	m.clients[c.ID] = c
	m.lock.Unlock()
}

func (m *broker) remove(c *phoenix.Server) {
	m.lock.Lock()
	delete(m.clients, c.ID)
	m.lock.Unlock()
}

func (m *broker) EmitAll(message *domain.Emit) {
	// clientCount := 0
	// mess := m.handleEmit(message)
	// for _, c := range m.clients {
	// 	if !c.connected {
	// 		m.remove(c)
	// 		break
	// 	}
	// 	for _, s := range c.subscribedTopics {
	// 		if s == mess.Topic {
	// 			err := c.Send(mess, false)
	// 			clientCount++
	// 			if err != nil {
	// 				velocity.GetLogger().Error("could not write message to client websocket", zap.Error(err), zap.String("clientID", c.ID))
	// 			}
	// 		}
	// 	}
	// }
}

func (m *broker) handleEmit(em *domain.Emit) *phoenix.PhoenixMessage {
	var payload interface{}
	var topic string

	switch v := em.Payload.(type) {
	case *user.User:
		break
	case *knownhost.KnownHost:
		topic = "knownhosts"
		payload = newKnownHostResponse(v)
		break
	case *project.Project:
		if em.Event == project.EventUpdate {
			topic = fmt.Sprintf("project:%s", v.Slug)
		} else {
			topic = "projects"
		}
		payload = newProjectResponse(v)
		break
	case *githistory.Branch:
		topic = fmt.Sprintf("project:%s", v.Project.Slug)
		payload = newBranchResponse(v)
		break
	case *githistory.Commit:
		topic = fmt.Sprintf("project:%s", v.Project.Slug)
		bs, _ := m.branchManager.GetAllForCommit(v, domain.NewPagingQuery())
		payload = newCommitResponse(v, bs)
		break
	// case *task.Task:
	// 	topic = fmt.Sprintf("project:%s", v.Commit.Project.Slug)
	// 	payload = newTaskResponse(v)
	// 	break
	case *build.Build:
		topic = fmt.Sprintf("project:%s", v.Task.Commit.Project.Slug)
		steps := m.stepManager.GetStepsForBuild(v)
		payload = newBuildResponse(v, stepsToStepResponse(steps, m.streamManager), m.branchManager)
		break
	case *build.Step:
		topic = fmt.Sprintf("project:%s", v.Build.Task.Commit.Project.Slug)
		steps := m.stepManager.GetStepsForBuild(v.Build)
		payload = newBuildResponse(v.Build, stepsToStepResponse(steps, m.streamManager), m.branchManager)
	case *build.StreamLine:
		topic = fmt.Sprintf("stream:%s", v.StreamID)
		payload = newStreamLineResponse(v)
		break
	default:
		velocity.GetLogger().Error("could not resolve websocket payload for client", zap.Any("payload", v))
	}

	// determine event
	var wsEvent string
	if val, ok := wsEventMapping[em.Event]; ok {
		wsEvent = val
	} else {
		velocity.GetLogger().Error("could not resolve websocket event for client", zap.Any("event", em.Event))
	}

	return &phoenix.PhoenixMessage{
		Topic:   topic,
		Event:   wsEvent,
		Payload: payload,
	}
}

var wsEventMapping = map[string]string{
	project.EventCreate: "project:new",
	project.EventUpdate: "project:update",
	project.EventDelete: "project:delete",

	githistory.EventCommitCreate: "commit:new",
	githistory.EventCommitUpdate: "commit:update",

	githistory.EventBranchCreate: "branch:new",
	githistory.EventBranchUpdate: "branch:update",

	knownhost.EventCreate: "knownhost:new",
	knownhost.EventDelete: "knownhost:delete",

	build.EventBuildCreate:      "build:new",
	build.EventBuildUpdate:      "build:update",
	build.EventStepUpdate:       "build:update",
	build.EventStreamLineCreate: "streamLine:new",

	// "": "builder:new",
	// "": "builder:update",
	// "": "builder:delete",
}
