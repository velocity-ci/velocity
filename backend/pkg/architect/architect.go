package architect

import (
	"context"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/velocity-ci/velocity/backend/pkg/architect/rest"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type architect struct {
	server   *echo.Echo
	workerWg sync.WaitGroup
	workers  []domain.Worker
}

func (a *architect) Start() {
	a.server.Start(":8080")
}

func (a *architect) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, w := range a.workers {
		w.StopWorker()
	}

	return a.server.Shutdown(ctx)
}

type App interface {
	Start()
	Stop() error
}

func New() App {
	a := &architect{
		server: echo.New(),
	}

	a.server.Use(middleware.Logger())
	a.server.Use(middleware.Recover())

	validator, trans := domain.NewValidator()
	db := domain.NewStormDB("architect.db")

	userManager := user.NewManager(db, validator, trans)
	knownHostManager := knownhost.NewManager(db, validator, trans)
	projectManager := project.NewManager(db, validator, trans, velocity.GitClone)
	commitManager := githistory.NewCommitManager(db)
	branchManager := githistory.NewBranchManager(db)
	taskManager := task.NewManager(db, projectManager, branchManager, commitManager)
	buildStepManager := build.NewStepManager(db)
	buildStreamFileManager := build.NewStreamFileManager(&a.workerWg, "/tmp/velocity-ci/logs")
	buildStreamManager := build.NewStreamManager(db, buildStreamFileManager)
	buildManager := build.NewBuildManager(db, buildStepManager, buildStreamManager)
	builderManager := builder.NewManager(buildManager, knownHostManager)

	rest.AddRoutes(
		a.server,
		userManager,
		knownHostManager,
		projectManager,
		commitManager,
		branchManager,
		taskManager,
		buildStepManager,
		buildStreamManager,
		buildManager,
		builderManager,
	)

	a.workers = []domain.Worker{
		builder.NewScheduler(builderManager, buildManager, &a.workerWg),
		buildStreamFileManager,
	}

	for _, w := range a.workers {
		go w.StartWorker()
	}

	return a
}
