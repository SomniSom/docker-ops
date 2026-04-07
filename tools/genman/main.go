// Genman writes man pages under man/man1/ (English UI strings). Run: go run ./tools/genman
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/SomniSom/docker-ops/internal/cli"
	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/version"
)

func main() {
	locale.Set("en")
	out := "man/man1"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	root := cli.NewRoot()
	header := &doc.GenManHeader{
		Section: "1",
		Source:  fmt.Sprintf("%s %s", version.Name, version.Version),
		Manual:  version.Name,
	}
	if err := doc.GenManTree(root, header, out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "wrote man pages under %s\n", out)
}
