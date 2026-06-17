// Command gw is the Groundwork CLI.
package main

import (
	"os"

	"groundwork/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args))
}
