package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
	"github.com/velocity-ci/velocity/backend/pkg/vcli"
)

var BuildVersion = "dev"

func main() {
	// vcli := vcli.New()

	// quit := make(chan os.Signal)
	// signal.Notify(quit, os.Interrupt)
	// go a.Start(quit)
	// <-quit
	// a.Stop()

	app := cli.NewApp()
	app.Name = "Velocity CLI"
	app.Usage = "Runs Velocity CI tasks locally"
	app.Version = BuildVersion

	app.Commands = []cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "List tasks",
			Action:  vcli.List,
		},
		// {
		// 	Name:    "run",
		// 	Aliases: []string{"r"},
		// 	Usage:   "Run a given task by name",
		// 	Action:  vcli.Run,
		// },
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "ignore-warnings",
			Usage: "ignore warnings during validation",
		},
		cli.BoolFlag{
			Name:  "ignore-errors",
			Usage: "ignore errors (and lower) during validation",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
