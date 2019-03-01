package builder

import (
	"fmt"
	"sync"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/builder"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"

	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"
)

type buildScheduler struct {
	builderManager *Manager
	stop           bool
	wg             *sync.WaitGroup

	buildManager     *build.BuildManager
	knownhostManager *knownhost.Manager
	stepManager      *build.StepManager
	streamManager    *build.StreamManager
}

func NewScheduler(
	wg *sync.WaitGroup,
	builderManager *Manager,
	buildManager *build.BuildManager,
	knownhostManager *knownhost.Manager,
	stepManager *build.StepManager,
	streamManager *build.StreamManager,
) *buildScheduler {
	return &buildScheduler{
		builderManager:   builderManager,
		knownhostManager: knownhostManager,
		stepManager:      stepManager,
		streamManager:    streamManager,
		buildManager:     buildManager,
		stop:             false,
		wg:               wg,
	}
}

func (bS *buildScheduler) StartWorker() {
	bS.wg.Add(1)
	// Requeue builds
	runningBuilds, _ := bS.buildManager.GetRunningBuilds()
	for _, runningBuild := range runningBuilds {
		runningBuild.Status = "waiting"
		bS.buildManager.Update(runningBuild)
	}
	velocity.GetLogger().Info("==> started build scheduler")
	for bS.stop == false {
		waitingBuilds, _ := bS.buildManager.GetWaitingBuilds()

		for _, waitingBuild := range waitingBuilds {
			// Queue on any idle worker
			activeBuilders, count := bS.builderManager.GetReady(domain.NewPagingQuery())
			velocity.GetLogger().Debug("got slaves", zap.Int("amount", count))
			for _, builder := range activeBuilders {
				bS.StartBuild(builder, waitingBuild)
				velocity.GetLogger().Info("starting build", zap.String("buildID", waitingBuild.ID), zap.String("builderID", builder.ID))
				break
			}
		}

		time.Sleep(1 * time.Second)
	}

	velocity.GetLogger().Info("==> stopped build scheduler")
	bS.wg.Done()
}

func (bS *buildScheduler) StopWorker() {
	bS.stop = true
}

func (bS *buildScheduler) StartBuild(bu *Builder, b *build.Build) {
	bu.State = StateBusy
	bS.builderManager.Save(bu)

	// Set knownhosts
	knownHosts, _ := bS.knownhostManager.GetAll(domain.NewPagingQuery())
	resp := bu.WS.Socket.Send(&phoenix.PhoenixMessage{
		// Event: builder.EventSetKnownHosts,
		Topic: fmt.Sprintf("builder:%s", bu.ID),
		Payload: &builder.KnownHostPayload{
			KnownHosts: knownHosts,
		},
	}, true)
	if resp.Status != phoenix.ResponseOK {
		velocity.GetLogger().Error("could not set knownhosts on builder", zap.String("builder", bu.ID))
	} else {
		velocity.GetLogger().Info("set knownhosts on builder", zap.String("builder", bu.ID))
	}

	// Start build
	b.Status = velocity.StateRunning
	bS.buildManager.Update(b)

	steps := bS.stepManager.GetStepsForBuild(b)
	streams := []*build.Stream{}
	for _, s := range steps {
		streams = append(streams, bS.streamManager.GetStreamsForStep(s)...)
	}

	resp = bu.WS.Socket.Send(&phoenix.PhoenixMessage{
		// Event: builder.EventStartBuild,
		Topic: fmt.Sprintf("builder:%s", bu.ID),
		// Payload: &builder.BuildPayload{
		// 	Build:   b,
		// 	Steps:   steps,
		// 	Streams: streams,
		// },
	}, true)
	if resp.Status != phoenix.ResponseOK {
		velocity.GetLogger().Error("could not start build on builder", zap.String("builder", bu.ID))
	} else {
		velocity.GetLogger().Info("started build on builder", zap.String("builder", bu.ID))
	}
}
