package rest

import (
	"os"

	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/domain/builder"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func AddRoutes(
	e *echo.Echo,
	userManager *user.Manager,
	knownHostManager *knownhost.Manager,
	projectManager *project.Manager,
	commitManager *githistory.CommitManager,
	branchManager *githistory.BranchManager,
	taskManager *task.Manager,
	buildStepManager *build.StepManager,
	buildStreamManager *build.StreamManager,
	buildManager *build.BuildManager,
	builderManager *builder.Manager,
) {
	// Unauthenticated routes
	authHandler := newAuthHandler(userManager)
	e.POST("/v1/auth", authHandler.create)

	// Authenticated routes
	userHandler := newUserHandler(userManager)
	knownHostHandler := newKnownHostHandler(knownHostManager)
	projectHandler := newProjectHandler(projectManager)
	commitHandler := newCommitHandler(projectManager, commitManager, branchManager)
	branchHandler := newBranchHandler(projectManager, branchManager, commitManager)
	taskHandler := newTaskHandler(projectManager, commitManager, branchManager, taskManager)
	buildHandler := newBuildHandler(buildManager, buildStepManager, buildStreamManager, projectManager, commitManager, branchManager, taskManager)
	buildStepHandler := newBuildStepHandler(buildManager, buildStepManager, buildStreamManager)
	buildStreamHandler := newBuildStreamHandler(buildStepManager, buildStreamManager)

	builderHandler := newBuilderHandler(builderManager)

	wsBroker := NewBroker(branchManager, buildStepManager, buildStreamManager)
	websocketHandler := newWebsocketHandler(wsBroker)
	userManager.AddBroker(wsBroker)
	knownHostManager.AddBroker(wsBroker)
	projectManager.AddBroker(wsBroker)
	commitManager.AddBroker(wsBroker)
	branchManager.AddBroker(wsBroker)
	taskManager.AddBroker(wsBroker)
	buildStepManager.AddBroker(wsBroker)
	buildStepManager.AddBroker(wsBroker)
	buildManager.AddBroker(wsBroker)
	buildStreamManager.AddBroker(wsBroker)

	// Used by Builders
	e.GET("/builder/ws", builderHandler.connect)

	jwtConfig := middleware.JWTConfig{
		Claims:        jwtStandardClaims,
		SigningKey:    []byte(os.Getenv("JWT_SECRET")),
		SigningMethod: jwtSigningMethod.Name,
	}

	r := e.Group("/v1/users")
	r.Use(middleware.JWTWithConfig(jwtConfig))
	r.POST("", userHandler.create)
	r.GET("", userHandler.getAll)
	r.GET("/:username", userHandler.get)
	r.DELETE("/:username", userHandler.delete)

	r = e.Group("/v1/ssh")
	r.Use(middleware.JWTWithConfig(jwtConfig))
	r.POST("/known-hosts", knownHostHandler.create)
	r.GET("/known-hosts", knownHostHandler.list)

	r = e.Group("/v1/projects")
	r.Use(middleware.JWTWithConfig(jwtConfig))
	r.POST("", projectHandler.create)
	r.GET("", projectHandler.getAll)
	r.GET("/:slug", projectHandler.get)
	r.POST("/:slug/sync", taskHandler.sync)

	r.GET("/:slug/branches", branchHandler.getAllForProject)
	r.GET("/:slug/branches/:name", branchHandler.getByProjectAndName)
	r.GET("/:slug/branches/:name/commits", branchHandler.getCommitsForBranch)
	r.GET("/:slug/commits", commitHandler.getAllForProject)
	r.GET("/:slug/commits/:hash", commitHandler.getByProjectAndHash)
	r.GET("/:slug/commits/:hash/tasks", taskHandler.getAllForCommit)
	r.GET("/:slug/commits/:hash/tasks/:taskSlug", taskHandler.getByProjectCommitAndSlug)

	r.POST("/:slug/commits/:hash/tasks/:taskSlug/builds", buildHandler.create)
	r.GET("/:slug/commits/:hash/tasks/:taskSlug/builds", buildHandler.getAllForTask)
	r.GET("/:slug/commits/:hash/builds", buildHandler.getAllForCommit)
	r.GET("/:slug/builds", buildHandler.getAllForProject)

	r = e.Group("/v1/builds")
	r.Use(middleware.JWTWithConfig(jwtConfig))
	r.GET("/:id", buildHandler.getByID)
	r.GET("/:id/steps", buildStepHandler.getStepsForBuildID)

	r = e.Group("/v1/steps")
	r.Use(middleware.JWTWithConfig(jwtConfig))
	r.GET("/:id", buildStepHandler.getByID)
	r.GET("/:id/streams", buildStreamHandler.getByStepID)

	r = e.Group("/v1/streams")
	r.Use(middleware.JWTWithConfig(jwtConfig))
	r.GET("/:id", buildStreamHandler.getByID)
	r.GET("/:id/log", buildStreamHandler.getLogByID)

	r = e.Group("/v1/builders")
	r.Use(middleware.JWTWithConfig(jwtConfig))
	r.GET("", builderHandler.getAll)
	r.GET("/:id", builderHandler.getByID)

	e.GET("/v1/ws", websocketHandler.phxClient)
}
