package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"
)

func newValidateCmd(projectDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: locale.T("validate.short"),
		Long:  locale.T("validate.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := config.ResolveProjectRoot(*projectDir)
			if err != nil {
				return err
			}
			p := config.FindDockerOpsFile(root)
			if p == "" {
				return fmt.Errorf("%s", locale.Tf("validate.err.no_docker_ops", root))
			}
			if err := config.ValidateFile(p); err != nil {
				return err
			}
			fmt.Fprint(cmd.OutOrStdout(), locale.Tf("msg.ok_path", p)+"\n")
			envPath := config.DQEnvPath(root)
			if st, err := os.Stat(envPath); err == nil && !st.IsDir() {
				if _, err := config.ParseDQEnv(envPath); err != nil {
					return fmt.Errorf("%s: %w", locale.T("validate.err.dq_env"), err)
				}
				fmt.Fprint(cmd.OutOrStdout(), locale.Tf("msg.ok_path", envPath)+"\n")
			}

			res, err := config.Load(root)
			if err != nil {
				return err
			}
			if res.Config.RemoteConfigured() {
				fmt.Fprint(cmd.ErrOrStderr(), locale.T("validate.note.remote"))
			}
			return nil
		},
	}
}
