package commands

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func getFile(c *cli.Context, name string) (*os.File, error) {
	inputFile := c.String(name)
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read input file '%s'. %w", name, err)
	}

	return file, nil
}
