package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/piotrjaromin/rolling-hash-algorithm/pkg/sync"
	"github.com/urfave/cli"
)

func NewSignatureCommand() cli.Command {
	return cli.Command{
		Name:  "signature",
		Usage: "Creates signature of a file",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:     "inputFile",
				Usage:    "File path for which signature should be calculated",
				Required: true,
			},
			cli.StringFlag{
				Name:     "signatureFile",
				Usage:    "File to which signature will be saved, if not provider it will be printed out",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			file, err := getFile(c, "inputFile")
			if err != nil {
				return err
			}
			defer file.Close()

			s := sync.New()

			chunkList := []sync.Chunk{}
			err = s.Signature(file, func(c sync.Chunk) {
				chunkList = append(chunkList, c)
			})

			if err != nil {
				return fmt.Errorf("error while calculating signature. %w", err)
			}

			serializedChunksReader, err := sync.SerializeChunks(chunkList)
			if err != nil {
				return fmt.Errorf("unable to serialize data chunks. %w", err)
			}

			serializedChunks, err := ioutil.ReadAll(serializedChunksReader)
			if err != nil {
				return fmt.Errorf("unable to read serialized data chunks. %w", err)
			}

			if c.IsSet("outputFile") {
				outputFile := c.String("outputFile")

				return os.WriteFile(outputFile, serializedChunks, os.ModePerm)
			}

			return nil
		},
	}
}
