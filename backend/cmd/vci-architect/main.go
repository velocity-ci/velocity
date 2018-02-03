package main

import (
	"os"
	"os/signal"

	"github.com/velocity-ci/velocity/backend/pkg/architect"
)

func main() {
	a := architect.New()

	go a.Start()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	a.Stop()
}