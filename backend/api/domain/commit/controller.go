package commit

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"github.com/velocity-ci/velocity/backend/api/auth"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
)

// Controller - Handles Commits.
type Controller struct {
	logger         *log.Logger
	render         *render.Render
	manager        Repository
	projectManager project.Repository
	// resolver       *Resolver
}

// NewController - Returns a new Controller for Commits.
func NewController(
	manager *Manager,
	projectManager *project.Manager,
	// commitResolver *Resolver,
) *Controller {
	return &Controller{
		logger:         log.New(os.Stdout, "[controller:commit]", log.Lshortfile),
		render:         render.New(),
		manager:        manager,
		projectManager: projectManager,
		// resolver:       commitResolver,
	}
}

// Setup - Sets up the routes for Commits.
func (c Controller) Setup(router *mux.Router) {
	// GET /v1/projects/{id}/commits
	router.Handle("/v1/projects/{id}/commits", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitsHandler)),
	)).Methods("GET")

	// GET /v1/projects/{projectID}/commits/{commitHash}
	router.Handle("/v1/projects/{projectID}/commits/{commitHash}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getProjectCommitHandler)),
	)).Methods("GET")

	// POST /v1/projects/{id}/commits/{commitHash}/builds
	// router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds", negroni.New(
	// 	auth.NewJWT(c.render),
	// 	negroni.Wrap(http.HandlerFunc(c.postProjectCommitBuildsHandler)),
	// )).Methods("POST")

	// GET /v1/projects/{id}/commits/{commitHash}/builds
	// router.Handle("/v1/projects/{projectID}/commits/{commitHash}/builds", negroni.New(
	// 	auth.NewJWT(c.render),
	// 	negroni.Wrap(http.HandlerFunc(c.getProjectCommitBuildsHandler)),
	// )).Methods("GET")

	c.logger.Println("Set up Commit controller.")
}

func (c Controller) getProjectCommitsHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["id"]

	p, err := c.projectManager.GetByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	opts := QueryOptsFromRequest(r)

	commits, count := c.manager.GetAllByProject(p, opts)

	respCommits := []*ResponseCommit{}
	for _, c := range commits {
		respCommits = append(respCommits, NewResponseCommit(c))
	}

	c.render.JSON(w, http.StatusOK, ManyResponse{
		Total:  count,
		Result: respCommits,
	})
}

func (c Controller) getProjectCommitHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitID := reqVars["commitHash"]

	p, err := c.projectManager.GetByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	commit, err := c.manager.GetByProjectAndHash(p, reqCommitID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	c.render.JSON(w, http.StatusOK, NewResponseCommit(commit))
}

// func (c Controller) postProjectCommitBuildsHandler(w http.ResponseWriter, r *http.Request) {
// 	reqVars := mux.Vars(r)
// 	reqProjectID := reqVars["projectID"]
// 	reqCommitID := reqVars["commitHash"]

// 	project, err := c.projectManager.FindByID(reqProjectID)
// 	if err != nil {
// 		log.Printf("Could not find project %s", reqProjectID)
// 		c.render.JSON(w, http.StatusNotFound, nil)
// 		return
// 	}

// 	commit, err := c.manager.GetCommitInProject(reqCommitID, project.ID)
// 	if err != nil {
// 		log.Printf("Could not find commit %s", reqCommitID)
// 		c.render.JSON(w, http.StatusNotFound, nil)
// 		return
// 	}

// 	build, err := c.resolver.BuildFromRequest(r.Body, project, commit)
// 	if err != nil {
// 		middleware.HandleRequestError(err, w, c.render)
// 		return
// 	}

// 	c.manager.SaveBuild(build, project.ID, commit.Hash)

// 	c.render.JSON(w, http.StatusCreated, NewResponseBuild(build, project.ID, commit.Hash))

// 	queuedBuild := NewQueuedBuild(build, project.ID, commit.Hash)
// 	c.manager.QueueBuild(queuedBuild)
// }

// func (c Controller) getProjectCommitBuildsHandler(w http.ResponseWriter, r *http.Request) {
// 	reqVars := mux.Vars(r)
// 	reqProjectID := reqVars["projectID"]
// 	reqCommitID := reqVars["commitHash"]

// 	project, err := c.projectManager.FindByID(reqProjectID)
// 	if err != nil {
// 		log.Printf("Could not find project %s", reqProjectID)
// 		c.render.JSON(w, http.StatusNotFound, nil)
// 		return
// 	}

// 	commit, err := c.manager.GetCommitInProject(reqCommitID, project.ID)
// 	if err != nil {
// 		log.Printf("Could not find commit %s", reqCommitID)
// 		c.render.JSON(w, http.StatusNotFound, nil)
// 		return
// 	}

// 	builds := c.manager.GetBuilds(project.ID, commit.Hash)

// 	respBuilds := []*ResponseBuild{}
// 	for _, b := range builds {
// 		respBuilds = append(respBuilds, NewResponseBuild(b, project.ID, commit.Hash))
// 	}

// 	c.render.JSON(w, http.StatusOK, respBuilds)
// }
