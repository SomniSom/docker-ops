package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/version"
)

func newManCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "man [command]...",
		Short: locale.T("man.short"),
		Long:  locale.T("man.long"),
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _, err := root.Find(args)
			if err != nil {
				return err
			}
			header := &doc.GenManHeader{
				Section: "1",
				Source:  fmt.Sprintf("%s %s", version.Name, version.Version),
				Manual:  version.Name,
			}
			var buf bytes.Buffer
			if err := doc.GenMan(target, header, &buf); err != nil {
				return fmt.Errorf("%s: %w", locale.T("man.err.gen"), err)
			}

			if manBin, err := exec.LookPath("man"); err == nil && manBin != "" {
				f, err := os.CreateTemp("", "dq-man-*.1")
				if err != nil {
					return err
				}
				path := f.Name()
				defer func() { _ = os.Remove(path) }()
				if _, err := f.Write(buf.Bytes()); err != nil {
					_ = f.Close()
					return err
				}
				if err := f.Close(); err != nil {
					return err
				}
				c := exec.Command(manBin, "-l", path)
				c.Stdin = os.Stdin
				c.Stdout = os.Stdout
				c.Stderr = os.Stderr
				if err := c.Run(); err != nil {
					return err
				}
				return nil
			}

			if _, err := os.Stdout.Write(buf.Bytes()); err != nil {
				return err
			}
			fmt.Fprintln(os.Stderr, locale.T("man.hint.no_man"))
			return nil
		},
	}
}
