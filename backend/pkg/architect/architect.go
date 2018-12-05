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
	v_sync "github.com/velocity-ci/velocity/backend/pkg/domain/sync"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type Architect struct {
	Server   *echo.Echo
	workerWg sync.WaitGroup
	Workers  []domain.Worker
	DB       *storm.DB
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

func New() *Architect {
	a := &Architect{
		Server: echo.New(),
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
	projectManager := project.NewManager(a.DB, validator, trans, velocity.Validate)
	commitManager := githistory.NewCommitManager(a.DB)
	branchManager := githistory.NewBranchManager(a.DB)
	taskManager := task.NewManager(a.DB, projectManager, branchManager, commitManager)
	buildStepManager := build.NewStepManager(a.DB)
	buildStreamManager := build.NewStreamManager(a.DB)
	buildManager := build.NewBuildManager(a.DB, buildStepManager, buildStreamManager)
	builderManager := builder.NewManager()
	syncManager := v_sync.NewManager(projectManager, taskManager, branchManager, commitManager)

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
		syncManager,
	)

	a.Workers = []domain.Worker{
		builder.NewScheduler(&a.workerWg, builderManager, buildManager, knownHostManager, buildStepManager, buildStreamManager),
	}
}
