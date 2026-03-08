package main

import (
	"fmt"
	"os"

	"github.com/barronlroth/f1-cli/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.Execute(version, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
