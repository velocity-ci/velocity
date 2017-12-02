package build

import (
	"fmt"
	"log"
	"os"

	"github.com/docker/go/canonical/json"
	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type GORMBuild struct {
	ID            string        `gorm:"primary_key"`
	Task          task.GORMTask `gorm:"ForeignKey:TaskReference"`
	TaskReference string
	Parameters    []byte // Parameters as JSON
	Status        string
}

type GORMBuildStep struct {
	ID             string    `gorm:"primary_key"`
	Build          GORMBuild `gorm:"ForeignKey:BuildReference"`
	BuildReference string
	Status         string
}

type GORMOutputStream struct {
	ID                 string        `gorm:"primary_key"`
	BuildStep          GORMBuildStep `gorm:"ForeignKey:BuildStepReference"`
	BuildStepReference string
	Name               string
	Path               string
}

func GORMBuildFromBuild(b Build) GORMBuild {
	jsonParameters, err := json.Marshal(b.Parameters)
	if err != nil {
		log.Printf("could not marshal build parameters from %v\n", b)
		log.Fatal(err)
	}
	return GORMBuild{
		ID:            b.ID,
		Task:          task.GORMTaskFromTask(b.Task),
		TaskReference: b.Task.ID,
		Parameters:    jsonParameters,
		Status:        b.Status,
	}
}

func BuildFromGORMBuild(gB GORMBuild) Build {
	var parameters map[string]velocity.Parameter
	err := json.Unmarshal(gB.Parameters, &parameters)
	if err != nil {
		log.Printf("could not unmarshal build parameters from %v\n", gB)
		log.Fatal(err)
	}
	return Build{
		ID:         gB.ID,
		Task:       task.TaskFromGORMTask(gB.Task),
		Parameters: parameters,
		Status:     gB.Status,
	}
}

func BuildStepFromGORMBuildStep(gBS GORMBuildStep) BuildStep {
	return BuildStep{
		ID:     gBS.ID,
		Status: gBS.Status,
		Build:  BuildFromGORMBuild(gBS.Build),
	}
}

func GORMBuildStepFromBuildStep(bS BuildStep) GORMBuildStep {
	return GORMBuildStep{
		ID:             bS.ID,
		Build:          GORMBuildFromBuild(bS.Build),
		BuildReference: bS.Build.ID,
		Status:         bS.Status,
	}
}

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	logger *log.Logger
	gorm   *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(GORMBuild{}, GORMBuildStep{})
	return &gormRepository{
		logger: log.New(os.Stdout, "[gorm:build]", log.Lshortfile),
		gorm:   db,
	}
}

func (r *gormRepository) SaveBuild(b Build) Build {
	tx := r.gorm.Begin()

	gormBuild := GORMBuildFromBuild(b)

	err := tx.Where(&GORMBuild{
		ID: b.ID,
	}).First(&GORMBuild{}).Error
	if err != nil {
		err = tx.Create(&gormBuild).Error
	} else {
		tx.Save(&gormBuild)
	}

	tx.Commit()

	return b

}
func (r *gormRepository) DeleteBuild(b Build) {
	tx := r.gorm.Begin()

	gormBuild := GORMBuildFromBuild(b)
	if err := tx.Delete(gormBuild).Error; err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	tx.Commit()
}

func (r *gormRepository) GetBuildByProjectAndCommitAndID(p project.Project, c commit.Commit, id string) (Build, error) {
	gormBuild := GORMBuild{}
	if r.gorm.
		Preload("Task").
		Preload("Task.Commit").
		Preload("Task.Commit.Project").
		Where(&GORMBuild{
			ID: id,
		}).
		First(&gormBuild).RecordNotFound() {
		r.logger.Printf("Could not find Build %s:%s:%s", p.ID, c.Hash, id)
		return Build{}, fmt.Errorf("could not find Build %s:%s:%s", p.ID, c.Hash, id)
	}
	return BuildFromGORMBuild(gormBuild), nil
}

func (r *gormRepository) GetBuildsByProject(p project.Project, q Query) ([]Build, uint64) {
	gormBuilds := []GORMBuild{}
	var count uint64
	r.gorm.
		Preload("Task").
		Preload("Task.Commit").
		Preload("Task.Commit.Project").
		Joins("JOIN gorm_tasks AS t ON t.id=gorm_builds.task_reference").
		Joins("JOIN commits AS c ON c.id=t.commit_reference").
		Joins("JOIN gorm_projects AS p ON p.id=c.project_reference").
		Where("p.id = ?", p.ID).
		Find(&gormBuilds).
		Count(&count)

	builds := []Build{}
	for _, gBuild := range gormBuilds {
		builds = append(builds, BuildFromGORMBuild(gBuild))
	}

	return builds, count
}

func (r *gormRepository) GetBuildsByProjectAndCommit(p project.Project, c commit.Commit) ([]Build, uint64) {
	gormBuilds := []GORMBuild{}
	var count uint64
	r.gorm.
		Preload("Task").
		Preload("Task.Commit").
		Preload("Task.Commit.Project").
		Joins("JOIN gorm_tasks AS t ON t.id=gorm_builds.task_reference").
		Joins("JOIN commits AS c ON c.id=t.commit_reference").
		Where("c.id = ?", c.ID).
		Find(&gormBuilds).
		Count(&count)

	builds := []Build{}
	for _, gBuild := range gormBuilds {
		builds = append(builds, BuildFromGORMBuild(gBuild))
	}

	return builds, count
}

func (r *gormRepository) GetRunningBuilds() ([]Build, uint64) {
	gormBuilds := []GORMBuild{}
	var count uint64
	r.gorm.
		Preload("Task").
		Preload("Task.Commit").
		Preload("Task.Commit.Project").
		Where(&GORMBuild{Status: "running"}).
		Find(&gormBuilds).
		Count(&count)

	builds := []Build{}
	for _, gBuild := range gormBuilds {
		builds = append(builds, BuildFromGORMBuild(gBuild))
	}

	return builds, count
}

func (r *gormRepository) GetWaitingBuilds() ([]Build, uint64) {
	gormBuilds := []GORMBuild{}
	var count uint64
	r.gorm.
		Preload("Task").
		Preload("Task.Commit").
		Preload("Task.Commit.Project").
		Where(&GORMBuild{Status: "waiting"}).
		Find(&gormBuilds).
		Count(&count)

	builds := []Build{}
	for _, gBuild := range gormBuilds {
		builds = append(builds, BuildFromGORMBuild(gBuild))
	}

	return builds, count
}

func (r *gormRepository) SaveBuildStep(bS BuildStep) BuildStep {
	tx := r.gorm.Begin()

	gormBuildStep := GORMBuildStepFromBuildStep(bS)

	err := tx.Where(&GORMBuildStep{
		ID: bS.ID,
	}).First(&GORMBuildStep{}).Error
	if err != nil {
		err = tx.Create(&gormBuildStep).Error
	} else {
		tx.Save(&gormBuildStep)
	}

	tx.Commit()

	return bS
}

func (r *gormRepository) GetBuildStepsForBuild(b Build) ([]BuildStep, uint64) {
	gormBuildSteps := []GORMBuildStep{}
	var count uint64

	r.gorm.
		Preload("Build").
		Preload("Build.Task").
		Preload("Build.Task.Commit").
		Preload("Build.Task.Commit.Project").
		Where(&GORMBuildStep{
			BuildReference: b.ID,
		}).Find(&gormBuildSteps).
		Count(&count)

	buildSteps := []BuildStep{}

	for _, gBuildStep := range gormBuildSteps {
		buildSteps = append(buildSteps, BuildStepFromGORMBuildStep(gBuildStep))
	}

	return buildSteps, count
}

func (r *gormRepository) GetBuildStepByBuildAndID(b Build, ID string) (BuildStep, error) {
	gormBuildStep := GORMBuildStep{}
	if r.gorm.
		Preload("Build").
		Preload("Build.Task").
		Where(&GORMBuildStep{
			ID:             ID,
			BuildReference: b.ID,
		}).
		First(&gormBuildStep).RecordNotFound() {
		r.logger.Printf("Could not find BuildStep %s:%s", b.ID, ID)
		return BuildStep{}, fmt.Errorf("could not find BuildStep %s:%s", b.ID, ID)
	}
	return BuildStepFromGORMBuildStep(gormBuildStep), nil
}

func (r *gormRepository) GetOutputStreamsForBuildStep(bS BuildStep) ([]velocity.OutputStream, uint64) {
	gormOutputStreams := []GORMOutputStream{}
	var count uint64

	r.gorm.
		Where(&GORMOutputStream{
			BuildStep: GORMBuildStep{
				ID: bS.ID,
			},
		}).Find(&gormOutputStreams).
		Count(&count)

	outputStreams := []velocity.OutputStream{}

	for _, gOutputStream := range gormOutputStreams {
		outputStreams = append(outputStreams, velocity.OutputStream{
			Name: gOutputStream.Name,
		})
	}

	return outputStreams, count
}

func (r *gormRepository) GetOutputStreamByID(id string) (velocity.OutputStream, error) {
	gormOutputStream := GORMOutputStream{}
	if r.gorm.
		Where(&GORMOutputStream{
			ID: id,
		}).
		First(&gormOutputStream).RecordNotFound() {
		r.logger.Printf("Could not find velocity.OutputStream %s", id)
		return velocity.OutputStream{}, fmt.Errorf("could not find velocity.OutputStream %s", id)
	}
	return velocity.OutputStream{
		Name: gormOutputStream.Name,
	}, nil
}
