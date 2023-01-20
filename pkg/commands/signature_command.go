package commands

import (
	"fmt"
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
			inputFile := c.String("inputFile")
			file, err := os.Open(inputFile)
			if err != nil {
				return fmt.Errorf("unable to read input file '%s'. %w", inputFile, err)
			}

			s := sync.New()

			chunkList := []sync.Chunk{}
			err = s.Signature(file, func(c sync.Chunk) {
				chunkList = append(chunkList, c)
			})

			if err != nil {
				return fmt.Errorf("error while calculating signature. %w", err)
			}

			serializedChunks := fmt.Sprintf("%+v", chunkList) // TODO improve

			if c.IsSet("outputFile") {
				outputFile := c.String("outputFile")
				return os.WriteFile(outputFile, []byte(serializedChunks), os.ModePerm)
			}

			fmt.Print(serializedChunks)
			return nil
		},
	}
}
