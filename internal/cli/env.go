package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"
)

func newEnvCmd(projectDir *string) *cobra.Command {
	var outPath string
	var force, anonymize bool

	cmd := &cobra.Command{
		Use:   "env",
		Short: locale.T("env.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := config.ResolveProjectRoot(*projectDir)
			if err != nil {
				return err
			}
			res, err := config.Load(root)
			if err != nil {
				return err
			}
			body := config.RemoteYAMLTemplate(res.Config, anonymize)
			if outPath == "" {
				_, err := fmt.Fprint(cmd.OutOrStdout(), body)
				return err
			}
			if _, err := os.Stat(outPath); err == nil && !force {
				return fmt.Errorf("refusing to overwrite %s (use --force)", outPath)
			}
			if err := os.WriteFile(outPath, []byte(body), 0o600); err != nil {
				return err
			}
			fmt.Fprint(os.Stderr, locale.Tf("env.wrote", outPath))
			return nil
		},
	}
	cmd.Flags().StringVarP(&outPath, "output", "o", "", "write template to file")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing output file")
	cmd.Flags().BoolVarP(&anonymize, "anonymize", "a", false, "do not substitute remote_ssh/remote_path from config")
	return cmd
}
