package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: locale.T("version.short"),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s (%s)\n", version.Name, version.Version, version.Commit)
		},
	}
}
