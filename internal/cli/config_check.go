package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SomniSom/docker-ops/internal/locale"
)

func newConfigCheckCmd(projectDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "config-check",
		Short: locale.T("configcheck.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, root, err := loadCfg(projectDir)
			if err != nil {
				return err
			}
			if cfg.AppConfig == "" {
				fmt.Fprintln(os.Stderr, "dq: app_config not set — nothing to check")
				return nil
			}
			if err := checkAppConfigExists(root, cfg.AppConfig); err != nil {
				return err
			}
			fmt.Print(locale.Tf("msg.ok_path", resolveAppConfigPath(root, cfg.AppConfig)) + "\n")
			return nil
		},
	}
}
