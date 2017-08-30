package routers

// MuxRouter - The application router
import (
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

// MuxRouter - A GorillaMux router.
type MuxRouter struct {
	Router  *mux.Router
	Negroni *negroni.Negroni
}

// Routable - Controllers should implement this.
type Routable interface {
	Setup(*mux.Router)
}

// NewMuxRouter - Sets up and returns a new MuxRouter with the given controllers.
func NewMuxRouter(controllers []Routable, logging bool) *MuxRouter {
	muxRouter := &MuxRouter{}

	muxRouter.Router = mux.NewRouter()

	muxRouter.Negroni = negroni.Classic()
	muxRouter.Negroni.Use(cors.New(cors.Options{
		AllowedHeaders: []string{
			"Authorization",
			"Content-Type",
		},
		AllowedMethods: []string{"GET", "POST", "PUT"},
		//Debug:          true,
	}))

	for _, controller := range controllers {
		controller.Setup(muxRouter.Router)
	}

	muxRouter.Negroni.UseHandler(muxRouter.Router)

	return muxRouter
}
