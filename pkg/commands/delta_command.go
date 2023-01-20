package commands

import (
	"fmt"

	"github.com/urfave/cli"
)

func NewDeltaCommand() cli.Command {
	return cli.Command{
		Name:  "delta",
		Usage: "",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:     "inputFile",
				Usage:    "File path for which delta should be calculated",
				Required: true,
			},
			cli.StringFlag{
				Name:     "signatureFile",
				Usage:    "Path to signature file which was calculated for previous version",
				Required: true,
			},
			cli.StringFlag{
				Name:     "deltaFile",
				Usage:    "File to which delta will be saved, if not provider it will be printed out",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			fmt.Println("DELTA")
			return nil
		},
	}
}
