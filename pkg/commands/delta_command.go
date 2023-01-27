package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/piotrjaromin/rolling-hash-algorithm/pkg/sync"
	"github.com/urfave/cli"
)

func NewDeltaCommand() cli.Command {
	return cli.Command{
		Name:  "delta",
		Usage: "Creates delta file with list of changes based in inputFile and signatureFile",
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
			file, err := getFile(c, "inputFile")
			if err != nil {
				return err
			}
			defer file.Close()

			sigFile, err := getFile(c, "signatureFile")
			if err != nil {
				return err
			}
			defer sigFile.Close()

			s := sync.New()

			deltas := []sync.Delta{}
			err = s.Delta(file, sigFile, func(d sync.Delta) {
				deltas = append(deltas, d)
			})

			if err != nil {
				return fmt.Errorf("error while calculating delta. %w", err)
			}

			serializedDeltaReader, err := sync.SerializeDeltas(deltas)
			if err != nil {
				return fmt.Errorf("unable to serialize deltas chunks. %w", err)
			}

			serializedDeltas, err := ioutil.ReadAll(serializedDeltaReader)
			if err != nil {
				return fmt.Errorf("unable to read serialized deltas chunks. %w", err)
			}

			if c.IsSet("deltaFile") {
				outputFile := c.String("deltaFile")
				return os.WriteFile(outputFile, []byte(serializedDeltas), os.ModePerm)
			}

			fmt.Printf("%+v", serializedDeltas)
			return nil
		},
	}
}
