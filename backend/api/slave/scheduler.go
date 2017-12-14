package slave

import (
	"log"
	"sync"
	"time"

	"github.com/velocity-ci/velocity/backend/api/domain/build"
)

type BuildScheduler struct {
	buildManager build.Repository
	slaveManager *Manager
	stop         bool
	wg           *sync.WaitGroup
}

func NewBuildScheduler(slaveManager *Manager, buildManager *build.Manager, wg *sync.WaitGroup) *BuildScheduler {
	return &BuildScheduler{
		slaveManager: slaveManager,
		buildManager: buildManager,
		stop:         false,
		wg:           wg,
	}
}

// TODO: Generate and persist BuildSteps and OutputStreams.
func (bS *BuildScheduler) StartWorker() {
	bS.wg.Add(1)
	// Requeue builds
	runningBuilds, _ := bS.buildManager.GetRunningBuilds()
	for _, runningBuild := range runningBuilds {
		runningBuild.Status = "waiting"
		bS.buildManager.UpdateBuild(runningBuild)
		log.Printf("Requeued: %s\n", runningBuild.ID)
	}
	log.Println("Started Build Scheduler")
	for bS.stop == false {
		waitingBuilds, total := bS.buildManager.GetWaitingBuilds()
		log.Printf("Got %d waiting builds", total)

		for _, waitingBuild := range waitingBuilds {
			log.Printf("%s: %s", waitingBuild.ID, waitingBuild.Status)
			// Queue on any idle worker
			activeSlaves, _ := bS.slaveManager.GetSlaves(SlaveQuery{Status: "ready"})
			for _, slave := range activeSlaves {
				go bS.slaveManager.StartBuild(slave, waitingBuild)
				break
			}
		}

		time.Sleep(10 * time.Second)
	}
	log.Println("Stopped Build Scheduler")
	bS.wg.Done()
}

func (bS *BuildScheduler) StopWorker() {
	bS.stop = true
}
