package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
	"github.com/velocity-ci/velocity/backend/pkg/vcli"
)

var BuildVersion = "dev"

func main() {

	app := cli.NewApp()
	app.Name = "Velocity CLI"
	app.Usage = "Runs Velocity CI tasks locally"
	app.Version = BuildVersion
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "List tasks",
			Action:  vcli.List,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "machine-readable",
					Usage: "Output in machine readable format (JSON)",
				},
			},
		},
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "Run a given task",
			Action:  vcli.Run,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "plan-only",
					Usage: "Only output the build plan",
				},
				cli.BoolFlag{
					Name:  "machine-readable",
					Usage: "Output in machine readable format (JSON)",
				},
			},
			// BashComplete: vcli.RunCompletion,
		},
		{
			Name:    "info",
			Aliases: []string{"i"},
			Usage:   "Print out information about the current project",
			Action:  vcli.Info,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "machine-readable",
					Usage: "Output in machine readable format (JSON)",
				},
			},
		},
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
