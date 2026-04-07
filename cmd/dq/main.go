// dq — Docker Quick-ops CLI (see readme.md).
package main

import (
	"fmt"
	"os"

	"github.com/SomniSom/docker-ops/internal/cli"
	"github.com/SomniSom/docker-ops/internal/locale"
)

func main() {
	locale.BootstrapFromArgs(os.Args[1:])
	if err := cli.NewRoot().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
