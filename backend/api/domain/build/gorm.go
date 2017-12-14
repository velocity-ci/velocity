package build

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/docker/go/canonical/json"
	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type gormBuild struct {
	ID          string `gorm:"primary_key"`
	TaskID      string
	Parameters  []byte // Parameters as JSON
	Status      string
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
}

func (gormBuild) TableName() string {
	return "builds"
}

type gormBuildStep struct {
	ID          string `gorm:"primary_key"`
	BuildID     string
	Number      uint64
	Status      string
	StartedAt   time.Time
	CompletedAt time.Time
}

func (gormBuildStep) TableName() string {
	return "build_steps"
}

type gormBuildStepStream struct {
	ID          string `gorm:"primary_key"`
	BuildStepID string
	Name        string
	Path        string
}

func (gormBuildStepStream) TableName() string {
	return "build_step_streams"
}

func gormBuildStepStreamFromBuildStepStream(s BuildStepStream) gormBuildStepStream {
	return gormBuildStepStream{
		ID:          s.ID,
		BuildStepID: s.BuildStepID,
		Name:        s.Name,
		Path:        s.ID, // TODO: define path in workspace
	}
}

func buildStepStreamFromGormBuildStepStream(g gormBuildStepStream) BuildStepStream {
	return BuildStepStream{
		ID:          g.ID,
		BuildStepID: g.BuildStepID,
		Name:        g.Name,
	}
}

func gormBuildFromBuild(b Build) gormBuild {
	jsonParameters, err := json.Marshal(b.Parameters)
	if err != nil {
		log.Printf("could not marshal build parameters from %v\n", b)
		log.Fatal(err)
	}
	return gormBuild{
		ID:          b.ID,
		TaskID:      b.TaskID,
		Parameters:  jsonParameters,
		Status:      b.Status,
		CreatedAt:   b.CreatedAt,
		StartedAt:   b.StartedAt,
		CompletedAt: b.CompletedAt,
	}
}

func buildFromGormBuild(g gormBuild) Build {
	var parameters map[string]velocity.Parameter
	err := json.Unmarshal(g.Parameters, &parameters)
	if err != nil {
		log.Printf("could not unmarshal build parameters from %v\n", g.Parameters)
		log.Fatal(err)
	}
	return Build{
		ID:          g.ID,
		TaskID:      g.TaskID,
		Parameters:  parameters,
		Status:      g.Status,
		CreatedAt:   g.CreatedAt,
		StartedAt:   g.StartedAt,
		CompletedAt: g.CompletedAt,
	}
}

func buildStepFromGormBuildStep(g gormBuildStep) BuildStep {
	return BuildStep{
		ID:      g.ID,
		BuildID: g.BuildID,
		Number:  g.Number,

		Status:      g.Status,
		StartedAt:   g.StartedAt,
		CompletedAt: g.CompletedAt,
	}
}

func gormBuildStepFromBuildStep(bS BuildStep) gormBuildStep {
	return gormBuildStep{
		ID:      bS.ID,
		BuildID: bS.BuildID,
		Number:  bS.Number,

		Status:      bS.Status,
		StartedAt:   bS.StartedAt,
		CompletedAt: bS.CompletedAt,
	}
}

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	logger *log.Logger
	gorm   *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(gormBuild{}, gormBuildStep{}, gormBuildStepStream{})
	return &gormRepository{
		logger: log.New(os.Stdout, "[gorm:build]", log.Lshortfile),
		gorm:   db,
	}
}

func (r *gormRepository) SaveBuild(b Build) Build {
	tx := r.gorm.Begin()

	gB := gormBuildFromBuild(b)

	err := tx.Where(&gormBuild{
		ID: b.ID,
	}).First(&gormBuild{}).Error
	if err != nil {
		err = tx.Create(&gB).Error
	} else {
		tx.Save(&gB)
	}

	tx.Commit()

	r.logger.Printf("saved build %s", b.ID)

	return buildFromGormBuild(gB)

}
func (r *gormRepository) DeleteBuild(b Build) {
	tx := r.gorm.Begin()

	gormBuild := gormBuildFromBuild(b)
	if err := tx.Delete(gormBuild).Error; err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	tx.Commit()
}

func (r *gormRepository) GetBuildByBuildID(buildID string) (Build, error) {
	gB := gormBuild{}
	if r.gorm.
		Where(&gormBuild{
			ID: buildID,
		}).
		First(&gB).RecordNotFound() {
		r.logger.Printf("could not find build %s", buildID)
		return Build{}, fmt.Errorf("could not find build %s", buildID)
	}
	return buildFromGormBuild(gB), nil
}

func (r *gormRepository) GetBuildsByProjectID(projectID string, q BuildQuery) ([]Build, uint64) {
	query := r.gorm.
		Joins("JOIN tasks AS t ON t.id=builds.task_id").
		Joins("JOIN commits AS c ON c.id=t.commit_id").
		Joins("JOIN projects AS p ON p.id=c.project_id").
		Where("p.id = ?", projectID)

	if q.Status != "all" {
		query = query.Where("builds.status = ?", q.Status)
	}

	return queryBuilds(query, q)
}

func (r *gormRepository) GetBuildsByCommitID(commitID string, q BuildQuery) ([]Build, uint64) {
	query := r.gorm.
		Joins("JOIN tasks AS t ON t.id=builds.task_id").
		Joins("JOIN commits AS c ON c.id=t.commit_id").
		Where("c.id = ?", commitID)

	if q.Status != "all" {
		query = query.Where(&gormBuild{Status: q.Status})
	}

	return queryBuilds(query, q)
}

func (r *gormRepository) GetBuildsByTaskID(taskID string, q BuildQuery) ([]Build, uint64) {
	query := r.gorm.
		Where(gormBuild{TaskID: taskID})

	if q.Status != "all" {
		query = query.Where(&gormBuild{Status: q.Status})
	}

	return queryBuilds(query, q)
}

func (r *gormRepository) GetRunningBuilds() ([]Build, uint64) {
	query := r.gorm.
		Where(&gormBuild{Status: "running"})

	return queryBuilds(query, BuildQuery{
		Amount: 10,
	})
}

func (r *gormRepository) GetWaitingBuilds() ([]Build, uint64) {
	query := r.gorm.
		Where(&gormBuild{Status: "waiting"})

	return queryBuilds(query, BuildQuery{
		Amount: 10,
	})
}

func queryBuilds(preparedDB *gorm.DB, q BuildQuery) ([]Build, uint64) {
	gBs := []gormBuild{}
	var count uint64
	preparedDB.
		Find(&gBs).
		Count(&count)
	preparedDB.
		Limit(int(q.Amount)).
		Offset(int(q.Page - 1)).
		Order("created_at desc").
		Find(&gBs)
	builds := []Build{}
	for _, gB := range gBs {
		builds = append(builds, buildFromGormBuild(gB))
	}

	return builds, count
}

func (r *gormRepository) SaveBuildStep(bS BuildStep) BuildStep {
	tx := r.gorm.Begin()

	gBS := gormBuildStepFromBuildStep(bS)

	err := tx.Where(&gormBuildStep{
		ID: bS.ID,
	}).First(&gormBuildStep{}).Error
	if err != nil {
		err = tx.Create(&gBS).Error
	} else {
		tx.Save(&gBS)
	}

	tx.Commit()

	return buildStepFromGormBuildStep(gBS)
}

func (r *gormRepository) DeleteBuildStep(bS BuildStep) {
	tx := r.gorm.Begin()

	gBS := gormBuildStepFromBuildStep(bS)
	if err := tx.Delete(gBS).Error; err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	tx.Commit()
	r.logger.Printf("removed buildStep %s", bS.ID)
}

func (r *gormRepository) GetBuildStepsByBuildID(buildID string) ([]BuildStep, uint64) {
	gBSs := []gormBuildStep{}
	var count uint64

	r.gorm.
		Where(&gormBuildStep{
			BuildID: buildID,
		}).Find(&gBSs).
		Count(&count)

	buildSteps := []BuildStep{}

	for _, gBS := range gBSs {
		buildSteps = append(buildSteps, buildStepFromGormBuildStep(gBS))
	}

	return buildSteps, count
}

func (r *gormRepository) GetBuildStepByBuildStepID(ID string) (BuildStep, error) {
	gBS := gormBuildStep{}
	if r.gorm.
		Where(&gormBuildStep{
			ID: ID,
		}).
		First(&gBS).RecordNotFound() {
		r.logger.Printf("could not find build step %s", ID)
		return BuildStep{}, fmt.Errorf("could not find build step %s", ID)
	}
	return buildStepFromGormBuildStep(gBS), nil
}

func (r *gormRepository) GetBuildStepByBuildIDAndNumber(buildID string, stepNumber uint64) (BuildStep, error) {
	gBS := gormBuildStep{}
	if r.gorm.
		Where(&gormBuildStep{
			BuildID: buildID,
			Number:  stepNumber,
		}).
		First(&gBS).RecordNotFound() {
		r.logger.Printf("could not find build step %s:%d", buildID, stepNumber)
		return BuildStep{}, fmt.Errorf("could not find build step %s:%d", buildID, stepNumber)
	}
	return buildStepFromGormBuildStep(gBS), nil
}

func (r *gormRepository) SaveStream(s BuildStepStream) BuildStepStream {
	tx := r.gorm.Begin()

	gS := gormBuildStepStreamFromBuildStepStream(s)

	err := tx.Where(&gormBuildStepStream{
		ID: s.ID,
	}).First(&gormBuildStepStream{}).Error
	if err != nil {
		err = tx.Create(&gS).Error
	} else {
		tx.Save(&gS)
	}

	tx.Commit()

	return buildStepStreamFromGormBuildStepStream(gS)
}

func (r *gormRepository) DeleteStream(s BuildStepStream) {
	tx := r.gorm.Begin()

	gS := gormBuildStepStreamFromBuildStepStream(s)
	if err := tx.Delete(gS).Error; err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	tx.Commit()
	r.logger.Printf("removed buildStepStream %s", s.ID)
}

func (r *gormRepository) GetStreamsByBuildStepID(buildStepID string) ([]BuildStepStream, uint64) {
	gBSSs := []gormBuildStepStream{}
	var count uint64

	r.gorm.
		Where(&gormBuildStepStream{
			BuildStepID: buildStepID,
		}).Find(&gBSSs).
		Count(&count)

	streams := []BuildStepStream{}

	for _, gBSS := range gBSSs {
		streams = append(streams, buildStepStreamFromGormBuildStepStream(gBSS))
	}

	return streams, count
}

func (r *gormRepository) GetStreamByID(id string) (BuildStepStream, error) {
	gBSS := gormBuildStepStream{}
	if r.gorm.
		Where(&gormBuildStepStream{
			ID: id,
		}).
		First(&gBSS).RecordNotFound() {
		r.logger.Printf("could not find build step stream %s", id)
		return BuildStepStream{}, fmt.Errorf("could not find build step stream %s", id)
	}
	return buildStepStreamFromGormBuildStepStream(gBSS), nil
}

func (r *gormRepository) GetStreamByBuildStepIDAndStreamName(buildStepID string, name string) (BuildStepStream, error) {
	gBSS := gormBuildStepStream{}
	if r.gorm.
		Where(&gormBuildStepStream{
			BuildStepID: buildStepID,
			Name:        name,
		}).
		First(&gBSS).RecordNotFound() {
		r.logger.Printf("could not find buildStep:stream %s:%s", buildStepID, name)
		return BuildStepStream{}, fmt.Errorf("could not find buildStep:stream %s:%s", buildStepID, name)
	}
	return buildStepStreamFromGormBuildStepStream(gBSS), nil
}
