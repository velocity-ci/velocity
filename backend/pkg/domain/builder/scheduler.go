package builder

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/velocity-ci/velocity/backend/pkg/domain"

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
	logrus.Info("Started Build scheduler")
	for bS.stop == false {
		waitingBuilds, total := bS.buildManager.GetWaitingBuilds()
		logrus.Debugf("Got %d waiting builds", total)

		for _, waitingBuild := range waitingBuilds {
			// Queue on any idle worker
			activeBuilders, count := bS.builderManager.GetReady(domain.NewPagingQuery())
			logrus.Debugf("Got %d ready slaves", count)
			for _, builder := range activeBuilders {
				go bS.builderManager.StartBuild(builder, waitingBuild)
				logrus.Infof("Starting build %s on %s", waitingBuild.ID, builder.ID)
				break
			}
		}

		time.Sleep(1 * time.Second)
	}
	logrus.Info("Stopped Build scheduler")
	bS.wg.Done()
}

func (bS *buildScheduler) StopWorker() {
	bS.stop = true
}
