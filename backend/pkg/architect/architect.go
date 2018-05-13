package architect

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/asdine/storm"
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
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type Architect struct {
	Server   *echo.Echo
	workerWg sync.WaitGroup
	Workers  []domain.Worker
	DB       *storm.DB
	LogsPath string
}

func (a *Architect) Start() {
	a.Init()
	a.Server.Use(middleware.Logger())
	a.Server.Use(middleware.Recover())
	a.Server.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	for _, w := range a.Workers {
		go w.StartWorker()
	}

	a.Server.Start(fmt.Sprintf(":%s", os.Getenv("PORT")))
}

func (a *Architect) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, w := range a.Workers {
		w.StopWorker()
	}

	return a.Server.Shutdown(ctx)
}

type App interface {
	Start()
	Stop() error
}

func New() *Architect {
	velocity.SetLogLevel()
	a := &Architect{
		Server:   echo.New(),
		LogsPath: "/opt/velocityci/logs",
	}

	return a
}

func (a *Architect) Init() {
	if a.DB == nil {
		a.DB = domain.NewStormDB("/opt/velocityci/architect.db")
	}
	validator, trans := domain.NewValidator()
	userManager := user.NewManager(a.DB, validator, trans)
	userManager.EnsureAdminUser()
	knownHostManager := knownhost.NewManager(a.DB, validator, trans, "")
	projectManager := project.NewManager(a.DB, validator, trans, velocity.Clone)
	commitManager := githistory.NewCommitManager(a.DB)
	branchManager := githistory.NewBranchManager(a.DB)
	taskManager := task.NewManager(a.DB, projectManager, branchManager, commitManager)
	buildStepManager := build.NewStepManager(a.DB)
	buildStreamFileManager := build.NewStreamFileManager(&a.workerWg, a.LogsPath)
	buildStreamManager := build.NewStreamManager(a.DB, buildStreamFileManager)
	buildManager := build.NewBuildManager(a.DB, buildStepManager, buildStreamManager)
	builderManager := builder.NewManager(buildManager, knownHostManager, buildStepManager, buildStreamManager)

	a.Server.Use(middleware.CORS())
	rest.AddRoutes(
		a.Server,
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

	a.Workers = []domain.Worker{
		builder.NewScheduler(builderManager, buildManager, &a.workerWg),
		buildStreamFileManager,
	}
}
