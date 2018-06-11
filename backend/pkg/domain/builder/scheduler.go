package builder

import (
	"sync"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"

	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
)

type buildScheduler struct {
	buildManager   *build.BuildManager
	builderManager *Manager
	stop           bool
	wg             *sync.WaitGroup
}

func NewScheduler(builderManager *Manager, buildManager *build.BuildManager, wg *sync.WaitGroup) *buildScheduler {
	return &buildScheduler{
		builderManager: builderManager,
		buildManager:   buildManager,
		stop:           false,
		wg:             wg,
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
				bS.builderManager.StartBuild(builder, waitingBuild)
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
