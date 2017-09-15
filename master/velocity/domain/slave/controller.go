package slave

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

type Controller struct {
	logger *log.Logger
	render *render.Render
}

func (c Controller) Setup(router *mux.Router) {
	router.
		HandleFunc("/v1/slave", c.postHandler).
		Methods("POST")
	c.logger.Println("Set up Slave controller.")
}

func (c Controller) postHandler(w http.ResponseWriter, r *http.Request) {

}
