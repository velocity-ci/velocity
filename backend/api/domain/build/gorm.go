package build

import (
	"log"

	"github.com/docker/go/canonical/json"
	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/project"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type GORMBuild struct {
	ID         string `gorm:"primary_key"`
	ProjectID  string
	CommitHash string
	TaskID     string
	Parameters []byte // Parameters as JSON
	Status     string
	BuildSteps []GORMBuildStep
}

type GORMBuildStep struct {
}

func gormBuildFromBuildAndProjectAndCommit(b *Build, p *project.Project, c *commit.Commit) *GORMBuild {
	jsonParameters, err := json.Marshal(b.Parameters)
	if err != nil {
		log.Fatal(err)
	}
	return &GORMBuild{
		ID:         b.ID,
		ProjectID:  p.ID,
		CommitHash: c.Hash,
		TaskID:     b.Task.ID,
		Parameters: jsonParameters,
		Status:     b.Status,
	}
}

func BuildFromGormBuildAndProjectAndCommitAndTask(gB *GORMBuild, p *project.Project, c *commit.Commit) *Build {
	var parameters []velocity.Parameter
	err := json.Unmarshal(&parameters)
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
func (r *gormRepository) Delete(b *Build) {

}
func (r *gormRepository) GetByProjectAndCommitAndID(p *project.Project, c *commit.Commit, id string) (*Build, error) {

}
func (r *gormRepository) GetAllByProject(p *project.Project, q Query) ([]*Build, uint64) {

}
func (r *gormRepository) GetAllByProjectAndCommit(p *project.Project, c *commit.Commit) ([]*Build, uint64) {

}
