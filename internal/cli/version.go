package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/version"
)

// newVersionCmd creates "version", printing binary name, semver, and commit hash.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: locale.T("version.short"),
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s %s (%s)\n", version.Name, version.Version, version.Commit)
			if version.ShowModuleVersionNote() {
				fmt.Fprintln(out, locale.T("version.hint_not_release"))
			}
		},
	}
}
