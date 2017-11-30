package slave

import (
	"log"
	"sync"
	"time"

	"github.com/velocity-ci/velocity/backend/api/domain/build"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
)

type BuildScheduler struct {
	commitManager  *commit.Manager
	buildManager   build.Repository
	slaveManager   *Manager
	projectManager *project.Manager
	stop           bool
	wg             *sync.WaitGroup
}

func NewBuildScheduler(commitManager *commit.Manager, slaveManager *Manager, projectManager *project.Manager, wg *sync.WaitGroup) *BuildScheduler {
	return &BuildScheduler{
		commitManager:  commitManager,
		slaveManager:   slaveManager,
		projectManager: projectManager,
		stop:           false,
		wg:             wg,
	}
}

// TODO: Generate and persist BuildSteps and OutputStreams.
func (bS *BuildScheduler) Run() {
	bS.wg.Add(1)
	// Requeue builds
	for _, runningBuild := range bS.buildManager.GetRunningBuilds() {
		runningBuild.Status = "waiting"
		bS.buildManager.SaveBuild(runningBuild)
		log.Printf("Requeued: %s\n", runningBuild.ID)
	}
	log.Println("Started Build Scheduler")
	for bS.stop == false {
		waitingBuilds := bS.buildManager.GetWaitingBuilds()
		log.Printf("Got %d waiting builds", len(queuedBuilds))

		for _, waitingBuild := range waitingBuilds {
			log.Printf("%s: %s", waitingBuild.ID, waitingBuild.Status)
			// Queue on any idle worker
			for _, slave := range bS.slaveManager.GetSlaves() {
				if slave.State == "ready" {
					go bS.slaveManager.StartBuild(slave, waitingBuild)
					break
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
	log.Println("Stopped Build Scheduler")
	bS.wg.Done()
}

func (bS *BuildScheduler) Stop() {
	bS.stop = true
}
