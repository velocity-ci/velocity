package build

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/go/canonical/json"
	"github.com/jinzhu/gorm"

	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
)

type GormBuild struct {
	UUID        string         `gorm:"primary_key"`
	Task        *task.GormTask `gorm:"ForeignKey:TaskID"`
	TaskID      string
	Parameters  []byte
	Status      string
	Steps       []*GormStep
	UpdatedAt   time.Time
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
}

func (GormBuild) TableName() string {
	return "builds"
}

func (g *GormBuild) ToBuild() *Build {
	params := map[string]string{}
	err := json.Unmarshal(g.Parameters, &params)
	if err != nil {
		logrus.Error(err)
	}

	steps := []*Step{}
	for _, g := range g.Steps {
		steps = append(steps, g.ToStep())
	}

	return &Build{
		UUID:        g.UUID,
		Task:        g.Task.ToTask(),
		Parameters:  params,
		Status:      g.Status,
		Steps:       steps,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
		StartedAt:   g.StartedAt,
		CompletedAt: g.CompletedAt,
	}
}

func (b *Build) ToGormBuild() *GormBuild {
	jsonParams, err := json.Marshal(b.Parameters)
	if err != nil {
		logrus.Error(err)
	}

	gSteps := []*GormStep{}
	for _, s := range b.Steps {
		gSteps = append(gSteps, s.ToGormStep())
	}

	return &GormBuild{
		UUID:        b.UUID,
		Task:        b.Task.ToGormTask(),
		Parameters:  jsonParams,
		Status:      b.Status,
		Steps:       gSteps,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
		StartedAt:   b.StartedAt,
		CompletedAt: b.CompletedAt,
	}
}

type buildDB struct {
	db *gorm.DB
}

func newBuildDB(gorm *gorm.DB) *buildDB {
	return &buildDB{
		db: gorm,
	}
}

func (db *buildDB) save(b *Build) error {
	tx := db.db.Begin()

	g := b.ToGormBuild()

	tx.
		Where(GormBuild{UUID: b.UUID}).
		Assign(&g).
		FirstOrCreate(&g)

	return tx.Commit().Error
}

func (db *buildDB) getAllForProject(p *project.Project, q *domain.PagingQuery) ([]*Build, int) {
	query := db.db.
		Joins("JOIN tasks AS t ON t.uuid=builds.task_id").
		Joins("JOIN commits AS c ON c.uuid=t.commit_id").
		Joins("JOIN projects AS p ON p.uuid=c.project_id").
		Where("p.uuid = ?", p.UUID)

	return queryBuilds(query, q)
}

func (db *buildDB) getAllForCommit(c *githistory.Commit, q *domain.PagingQuery) ([]*Build, int) {
	query := db.db.
		Joins("JOIN tasks AS t ON t.uuid=builds.task_id").
		Joins("JOIN commits AS c ON c.uuid=t.commit_id").
		Where("c.uuid = ?", c.UUID)

	return queryBuilds(query, q)
}

func (db *buildDB) getAllForBranch(b *githistory.Branch, q *domain.PagingQuery) ([]*Build, int) {
	query := db.db.
		Joins("JOIN tasks AS t ON t.uuid=builds.task_id").
		Joins("JOIN commits AS c ON c.uuid=t.commit_id").
		Joins("JOIN commit_branches AS cb ON cb.gorm_commit_uuid=c.uuid").
		Joins("JOIN branches AS b ON b.uuid=cb.gorm_branch_uuid").
		Where("b.uuid = ?", b.UUID)

	return queryBuilds(query, q)
}

func (db *buildDB) getAllForTask(t *task.Task, q *domain.PagingQuery) ([]*Build, int) {
	query := db.db.
		Joins("JOIN tasks AS t ON t.uuid=builds.task_id").
		Where("t.uuid = ?", t.UUID)

	return queryBuilds(query, q)
}

func (db *buildDB) getRunningBuilds() ([]*Build, int) {
	query := db.db.
		Where("status = ?", "running")

	return queryBuilds(query, &domain.PagingQuery{
		Limit: 100,
		Page:  1,
	})
}

func (db *buildDB) getWaitingBuilds() ([]*Build, int) {
	query := db.db.
		Where("status = ?", "waiting")

	return queryBuilds(query, &domain.PagingQuery{
		Limit: 100,
		Page:  1,
	})
}

func queryBuilds(preparedDB *gorm.DB, q *domain.PagingQuery) (r []*Build, t int) {
	t = 0
	gBs := []*GormBuild{}
	preparedDB.
		Find(&gBs).
		Count(&t)
	preparedDB.LogMode(true).
		// Preload("Task").
		// Preload("Task.Commit").
		// Preload("Task.Commit.Project").
		// Preload("Task.Commit.Branches").
		// Preload("Task.Commit.Branches.Project").
		Joins("JOIN tasks AS t ON t.uuid=builds.task_id").
		Joins("JOIN commits AS c ON c.uuid=t.commit_id").
		Joins("JOIN projects AS cp ON cp.uuid=c.project_id").
		Joins("JOIN commit_branches AS cb ON cb.gorm_commit_uuid=c.uuid").
		Joins("JOIN branches AS b ON b.uuid=cb.gorm_branch_uuid").
		Joins("JOIN projects AS bp ON bp.uuid=b.project_id").
		// Preload("Steps").
		// Preload("Steps.Streams").
		Limit(int(q.Limit)).
		Offset(int(q.Page - 1)).
		Order("created_at desc").
		Find(&gBs)
	for _, gB := range gBs {
		r = append(r, gB.ToBuild())
	}

	return r, t
}

func (db *buildDB) getBuildByUUID(uuid string) (*Build, error) {
	g := GormBuild{}
	if db.db.Preload("Task").
		Preload("Task.Commit").
		Preload("Task.Commit.Project").
		Preload("Task.Commit.Branches").
		Preload("Task.Commit.Branches.Project").
		Where("uuid = ?", uuid).First(&g).RecordNotFound() {
		return nil, fmt.Errorf("could not find build %s", uuid)
	}

	return g.ToBuild(), nil
}
