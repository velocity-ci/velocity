package main

import (
	"os"
	"os/signal"

	"github.com/velocity-ci/velocity/backend/architect/web"
)

func main() {
	webApp := web.NewWeb()

	go webApp.Start()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	webApp.Stop()
}
