package rest

import (
	"os"

	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"

	"github.com/velocity-ci/velocity/backend/velocity"

	"github.com/go-playground/universal-translator"
	"github.com/jinzhu/gorm"
	validator "gopkg.in/go-playground/validator.v9"

	"github.com/velocity-ci/velocity/backend/pkg/domain/knownhost"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func AddRoutes(
	e *echo.Echo,
	db *gorm.DB,
	validator *validator.Validate,
	trans ut.Translator,
) {
	// Unauthenticated routes
	userManager := user.NewManager(db, validator, trans)
	authHandler := newAuthHandler(userManager)
	e.POST("/v1/auth", authHandler.create)

	// Authenticated routes
	knownHostManager := knownhost.NewManager(db, validator, trans)
	knownHostHandler := newKnownHostHandler(knownHostManager)
	projectManager := project.NewManager(db, validator, trans, velocity.GitClone)
	projectHandler := newProjectHandler(projectManager)
	commitManager := githistory.NewCommitManager(db)
	commitHandler := newCommitHandler(projectManager, commitManager)
	branchManager := githistory.NewBranchManager(db)
	branchHandler := newBranchHandler(projectManager, branchManager)
	taskManager := task.NewManager(db)
	taskHandler := newTaskHandler(projectManager, commitManager, taskManager)

	jwtConfig := middleware.JWTConfig{
		Claims:     &jwt.StandardClaims{},
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
	}

	r := e.Group("/v1/ssh")
	r.Use(middleware.JWTWithConfig(jwtConfig))
	r.POST("/known-hosts", knownHostHandler.create)

	r = e.Group("/v1/projects")
	r.Use(middleware.JWTWithConfig(jwtConfig))
	r.POST("", projectHandler.create)
	r.GET("", projectHandler.getAll)
	r.GET("/:slug", projectHandler.get)

	r.GET("/:slug/branches", branchHandler.getAll)
	r.GET("/:slug/commits", commitHandler.getAllForProject)
	r.GET("/:slug/commits/:hash", commitHandler.getByProjectAndHash)
	r.GET("/:slug/commits/:hash/tasks", taskHandler.getAllForCommit)

}
