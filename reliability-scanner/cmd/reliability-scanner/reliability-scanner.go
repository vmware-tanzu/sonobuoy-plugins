package main

import (
	"os"
	"log"
	"github.com/urfave/cli/v2"
)

var (
	reportName = "reliability-scanner"
	filePath   = "/tmp/results/reliability.yaml"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:    "scan",
				Aliases: []string{"s"},
				Usage:   "run the scanner",
				Action: func(c *cli.Context) error {
					scan()
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
