package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	var returnCode = make(chan int)
	var finishUP = make(chan struct{})
	var done = make(chan struct{})

	var gracefulStop = make(chan os.Signal)

	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		// wait for our os signal to stop the app
		// on the graceful stop channel
		sig := <-gracefulStop
		log.Printf("Caught signal: %+s\n", sig)

		// send message on "finish up" channel to tell the app to
		// gracefully shutdown
		finishUP <- struct{}{}

		// wait for word back if we finished or not
		select {
		case <-time.After(30 * time.Second):
			// timeout after 30 seconds waiting for app to finish
			returnCode <- 1
		case <-done:
			// if we got a message on done, we finished, so end app
			returnCode <- 0
		}
	}()

	var app App

	// if os.Getenv("TYPE") == "API" {
	app = NewVelocityAPI()
	// } else if os.Getenv("TYPE") == "WORKER" {
	// app = NewPrivNegWorker(os.Getenv("QUEUE"))
	// } else {
	// panic(fmt.Sprintf("Invalid backend type: %s. Must be api|worker", os.Getenv("TYPE")))
	// }
	go app.Start()

	fmt.Println("Waiting for finish")
	// wait for finishUP channel write to close the app down
	<-finishUP
	fmt.Println("Stopping Privacy Negotiator App")
	app.Stop()
	// write to the done channel to signal we are done.
	done <- struct{}{}
	os.Exit(<-returnCode)
}

// App - For starting and stopping gracefully.
type App interface {
	Start()
	Stop()
}
