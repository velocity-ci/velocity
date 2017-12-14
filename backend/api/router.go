package main

import (
	"fmt"
	"log"
	"strings"

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
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		// Debug:          true,
	}))

	for _, controller := range controllers {
		controller.Setup(muxRouter.Router)
	}

	routes := []string{}

	muxRouter.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		m, _ := route.GetMethods()
		routes = append(routes, fmt.Sprintf("%s: %s", m[0], t))
		return nil
	})

	log.Printf("\n\n%s\n\n", strings.Join(routes, "\n"))

	muxRouter.Negroni.UseHandler(muxRouter.Router)

	return muxRouter
}
