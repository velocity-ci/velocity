package slave

import (
	"log"
	"sync"
	"time"

	"github.com/velocity-ci/velocity/backend/api/commit"
	"github.com/velocity-ci/velocity/backend/api/project"
)

type BuildScheduler struct {
	commitManager  *commit.Manager
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

func (bS *BuildScheduler) Run() {
	bS.wg.Add(1)
	// Requeue builds
	for _, queuedBuild := range bS.commitManager.GetQueuedBuilds() {
		build := bS.commitManager.GetBuildFromQueuedBuild(queuedBuild)
		log.Printf("Reset: %s, %s, %d", queuedBuild.ProjectID, queuedBuild.CommitHash, queuedBuild.ID)
		build.Status = "waiting"
		bS.commitManager.SaveBuild(build, queuedBuild.ProjectID, queuedBuild.CommitHash)
	}
	log.Println("Started Build Scheduler")
	for bS.stop == false {
		queuedBuilds := bS.commitManager.GetQueuedBuilds()
		log.Printf("Got %d queued builds", len(queuedBuilds))

		for _, queuedBuild := range queuedBuilds {
			build := bS.commitManager.GetBuildFromQueuedBuild(queuedBuild)
			log.Printf("%s:%s:%d: %s", queuedBuild.ProjectID, queuedBuild.CommitHash, queuedBuild.ID, build.Status)
			if build.Status == "waiting" {
				// Queue on any idle worker
				for _, slave := range bS.slaveManager.GetSlaves() {
					if slave.State == "ready" {
						project, _ := bS.projectManager.FindByID(queuedBuild.ProjectID)
						// convert queued build to build
						go bS.slaveManager.StartBuild(slave.ID, project, queuedBuild.CommitHash, build)
						build.Status = "running"
						bS.commitManager.SaveBuild(build, project.ID, queuedBuild.CommitHash)
						break
					}
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
