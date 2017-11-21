package build

import (
	"fmt"
	"log"

	"github.com/docker/go/canonical/json"
	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type GORMBuild struct {
	ID               string              `gorm:"primary_key"`
	Project          project.GORMProject `gorm:"ForeignKey:ProjectReference"`
	ProjectReference string
	Commit           commit.GORMCommit `gorm:"ForeignKey:CommitReference"`
	CommitReference  string
	Task             task.GORMTask `gorm:"ForeignKey:TaskReference"`
	TaskReference    string
	Parameters       []byte // Parameters as JSON
	Status           string
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

func gormBuildFromBuild(b *Build) *GORMBuild {
	jsonParameters, err := json.Marshal(b.Parameters)
	if err != nil {
		log.Fatal(err)
	}
	return &GORMBuild{
		ID:               b.ID,
		Project:          project.GORMProjectFromProject(b.Project),
		ProjectReference: b.Project.ID,
		Commit:           commit.GORMCommitFromCommit(b.Commit),
		CommitReference:  b.Commit.ID,
		Task:             task.GORMTaskFromTask(b.Task),
		TaskReference:    b.Task.ID,
		Parameters:       jsonParameters,
		Status:           b.Status,
	}
}

func buildFromGormBuildAndProjectAndCommitAndTask(gB *GORMBuild, p *project.Project, c *commit.Commit, t *task.Task) *Build {
	var parameters []velocity.Parameter
	err := json.Unmarshal(gB.Parameters, &parameters)
	if err != nil {
		log.Fatal(err)
	}
	return &Build{
		ID:         gB.ID,
		Project:    *p,
		Commit:     *c,
		Task:       *t,
		Parameters: parameters,
		Status:     gB.Status,
	}
}

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	gorm *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(GORMBuild{})
	return &gormRepository{
		gorm: db,
	}
}

func (r *gormRepository) SaveToProjectAndCommit(p *project.Project, c *commit.Commit, b *Build) *Build {
	tx := r.gorm.Begin()

	gormBuild := gormBuildFromBuildAndProjectAndCommit(b, p, c)

	tx.
		Where(GORMBuild{
			ID:         gormBuild.ID,
			ProjectID:  gormBuild.ProjectID,
			CommitHash: gormBuild.CommitHash,
			TaskID:     gormBuild.TaskID,
		}).
		Assign(gormBuild).
		FirstOrCreate(gormBuild)

	tx.Commit()

	return b

}
func (r *gormRepository) DeleteFromProjectAndCommit(p *project.Project, c *commit.Commit, b *Build) {
	tx := r.gorm.Begin()

	gormBuild := gormBuildFromBuildAndProjectAndCommit(b, p, c)
	if err := tx.Delete(gormBuild).Error; err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	tx.Commit()
}

func (r *gormRepository) GetByProjectAndCommitAndID(p *project.Project, c *commit.Commit, id string) (*Build, error) {
	gormBuild := &GORMBuild{}
	if r.gorm.
		Where(&GORMBuild{ProjectID: p.ID, CommitHash: c.Hash, ID: id}).
		First(gormBuild).RecordNotFound() {
		log.Printf("Could not find Build %s:%s:%s", p.ID, c.Hash, id)
		return nil, fmt.Errorf("could not find Build %s:%s:%s", p.ID, c.Hash, id)
	}

	t := task.Task{}
	r.gorm.
		Where(&task.GORMTask{ID: gormBuild.TaskID, CommitHash: c.Hash, ProjectID: p.ID}).
		First(&t)

	return buildFromGormBuildAndProjectAndCommitAndTask(gormBuild, p, c, &t), nil
}

func (r *gormRepository) GetAllByProject(p *project.Project, q Query) ([]*Build, uint64) {
	gormBuilds := []GORMBuild{}
	var count uint64
	r.gorm.
		Where(&GORMBuild{ProjectID: p.ID}).
		Find(&gormBuilds).
		Count(&count)

	builds := []*Build{}
	for _, gBuild := range gormBuilds {
		c := commit.Commit{}
		r.gorm.
			Where(&commit.GORMCommit{ProjectID: p.ID, Hash: gBuild.CommitHash}).
			First(&c)
		t := task.Task{}
		r.gorm.
			Where(&task.GORMTask{ID: gBuild.TaskID, ProjectID: p.ID, CommitHash: c.Hash}).
			First(&t)

		builds = append(builds, buildFromGormBuildAndProjectAndCommitAndTask(&gBuild, p, &c, &t))
	}

	return builds, count
}

func (r *gormRepository) GetAllByProjectAndCommit(p *project.Project, c *commit.Commit) ([]*Build, uint64) {
	gormBuilds := []GORMBuild{}
	var count uint64
	r.gorm.
		Where(&GORMBuild{ProjectID: p.ID, CommitHash: c.Hash}).
		Find(&gormBuilds).
		Count(&count)

	builds := []*Build{}
	for _, gBuild := range gormBuilds {
		t := task.Task{}
		r.gorm.
			Where(&task.GORMTask{ID: gBuild.TaskID, ProjectID: p.ID, CommitHash: c.Hash}).
			First(&t)

		builds = append(builds, buildFromGormBuildAndProjectAndCommitAndTask(&gBuild, p, c, &t))
	}

	return builds, count
}

func (r *gormRepository) SaveBuildStep(bS *BuildStep) *BuildStep {
	tx := r.gorm.Begin()

	gormBuildStep := &GORMBuildStep{
		ID:             bS.ID,
		Build:          gormBuildFromBuild(bS.Build),
		BuildReference: bS.Build.ID,
		Status:         bS.Status,
	}

	tx.
		Where(GORMBuildStep{
			ID: gormBuild.ID,
		}).
		Assign(gormBuildStep).
		FirstOrCreate(gormBuildStep)

	tx.Commit()

	return b

}

func (r *gormRepository) GetBuildStepsForBuild(b *Build) ([]*BuildStep, uint64) {
	gormBuildSteps := []GORMBuildStep{}
	var count uint64

	r.gorm.
		Where(&GORMBuildStep{
			Build: GORMBuild{
				ID:         b.ID,
				ProjectID:  b.Project.ID,
				CommitHash: b.Commit.Hash,
			},
		}).Find(&gormBuildSteps).
		Count(&count)

	buildSteps := []*BuildStep{}

	for _, gBuildStep := range gormBuildSteps {
		buildSteps = append(buildSteps, &BuildStep{
			ID:     gBuildStep.ID,
			Build:  *b,
			Status: gBuildStep.Status,
		})
	}

	return buildSteps, count
}

func (r *gormRepository) GetOutputStreamsForBuildStep(bS *BuildStep) ([]*OutputStream, uint64) {
	gormOutputStreams := []GORMOutputStream{}
	var count uint64

	r.gorm.
		Where(&GORMOutputStream{
			BuildStep: GORMBuildStep{
				ID: bS.ID,
			},
		}).Find(&gormOutputStreams).
		Count(&count)

	outputStreams := []*OutputStream{}

	for _, gOutputStream := range gormOutputStreams {
		outputStreams = append(outputStreams, &OutputStream{
			ID:        gOutputStream.ID,
			BuildStep: *bS,
			Name:      gOutputStream.Name,
			Path:      gOutputStream.Path,
		})
	}

	return outputStreams, count
}
