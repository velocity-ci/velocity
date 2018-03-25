package rest

import (
	"fmt"
	"sync"

	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"

	"github.com/Sirupsen/logrus"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
)

type broker struct {
	clients map[string]*Client
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
		clients:       map[string]*Client{},
		lock:          sync.RWMutex{},
		branchManager: branchManager,
		stepManager:   stepManager,
		streamManager: streamManager,
	}
}

func (m *broker) save(c *Client) {
	m.lock.Lock()
	m.clients[c.ID] = c
	m.lock.Unlock()
}

func (m *broker) remove(c *Client) {
	m.lock.Lock()
	delete(m.clients, c.ID)
	m.lock.Unlock()
}

func (m *broker) monitor(c *Client) {
	c.alive = true
	for {
		message := &PhoenixMessage{}
		err := c.ws.ReadJSON(message)
		if err != nil {
			c.alive = false
			logrus.Error(err)
			logrus.Infof("Closing Client WebSocket: %s", c.ID)
			c.ws.Close()
			m.remove(c)
			return
		}
		c.HandleMessage(message)
		m.save(c)
	}
}

func (m *broker) EmitAll(message *domain.Emit) {
	clientCount := 0
	mess := m.handleEmit(message)
	for _, c := range m.clients {
		if !c.alive {
			m.remove(c)
			break
		}
		for _, s := range c.subscriptions {
			if s == mess.Topic {
				err := c.WriteJSON(mess)
				clientCount++
				if err != nil {
					logrus.Println(err)
				}
			}
		}
	}
}

func (m *broker) handleEmit(em *domain.Emit) *PhoenixMessage {
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
		logrus.Errorf("could not resolve websocket payload %+v", v)
	}

	// determine event
	var wsEvent string
	if val, ok := wsEventMapping[em.Event]; ok {
		wsEvent = val
	} else {
		logrus.Errorf("could not resolve event in websocket: %s", em.Event)
	}

	return &PhoenixMessage{
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
	build.EventStepUpdate:       "step:update",
	build.EventStreamLineCreate: "streamLine:new",

	// "": "builder:new",
	// "": "builder:update",
	// "": "builder:delete",
}
