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

	// GET /v1/commits/{commitUUID}
	router.Handle("/v1/commits/{commitUUID}", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getCommitByUUIDHandler)),
	)).Methods("GET")

	// GET /v1/projects/{id}/branches
	router.Handle("/v1/projects/{id}/branches", negroni.New(
		auth.NewJWT(c.render),
		negroni.Wrap(http.HandlerFunc(c.getBranchesHandler)),
	)).Methods("GET")

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

	opts := CommitQueryOptsFromRequest(r)

	commits, count := c.manager.GetAllCommitsByProjectID(p.ID, opts)

	respCommits := []ResponseCommit{}
	for _, c := range commits {
		respCommits = append(respCommits, NewResponseCommit(c))
	}

	c.render.JSON(w, http.StatusOK, ManyResponseCommit{
		Total:  count,
		Result: respCommits,
	})
}

func (c Controller) getCommitByUUIDHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqCommitUUID := reqVars["commitUUID"]

	commit, err := c.manager.GetCommitByCommitID(reqCommitUUID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	c.render.JSON(w, http.StatusOK, NewResponseCommit(commit))
}

func (c Controller) getProjectCommitHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["projectID"]
	reqCommitHash := reqVars["commitHash"]

	p, err := c.projectManager.GetByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	commit, err := c.manager.GetCommitByProjectIDAndCommitHash(p.ID, reqCommitHash)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	c.render.JSON(w, http.StatusOK, NewResponseCommit(commit))
}

func (c Controller) getBranchesHandler(w http.ResponseWriter, r *http.Request) {
	reqVars := mux.Vars(r)
	reqProjectID := reqVars["id"]

	p, err := c.projectManager.GetByID(reqProjectID)
	if err != nil {
		c.render.JSON(w, http.StatusNotFound, nil)
		return
	}

	opts := BranchQueryOptsFromRequest(r)

	branches, count := c.manager.GetAllBranchesByProjectID(p.ID, opts)

	respBranches := []ResponseBranch{}

	for _, b := range branches {
		respBranches = append(respBranches, NewResponseBranch(b))
	}

	c.render.JSON(w, http.StatusOK, ManyResponseBranch{
		Total:  count,
		Result: respBranches,
	})
}
