package main

import (
	"log"
	"os"

	"github.com/piotrjaromin/rolling-hash-algorithm/pkg/commands"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		commands.NewDeltaCommand(),
		commands.NewSignatureCommand(),
	}

	app.Name = "App for calculating hashes and deltas of files"
	app.Version = "1.0.0"

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
