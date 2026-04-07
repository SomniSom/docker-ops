package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SomniSom/docker-ops/internal/deploy"
	"github.com/SomniSom/docker-ops/internal/locale"
)

// newDeployCmd creates "deploy": requires remote_ssh/remote_path in config, then runs
// deploy.RunWithOptions (optional --build for image artifacts).
func newDeployCmd(projectDir *string) *cobra.Command {
	var build bool
	c := &cobra.Command{
		Use:   "deploy",
		Short: locale.T("deploy.short"),
		Long:  locale.T("deploy.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, root, err := loadCfg(projectDir)
			if err != nil {
				return err
			}
			if !cfg.RemoteConfigured() {
				return fmt.Errorf("%s", locale.T("deploy.err.needs_remote"))
			}
			return deploy.RunWithOptions(root, cfg, deploy.RunOpts{Build: build})
		},
	}
	c.Flags().BoolVar(&build, "build", false, locale.T("deploy.flag.build"))
	return c
}
