package cli

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SomniSom/docker-ops/internal/composeimage"
	"github.com/SomniSom/docker-ops/internal/locale"
)

// newGenImageComposeCmd creates "gen-image-compose": emits a compose override with a
// pinned image reference for deploy workflows (see composeimage package and flags).
func newGenImageComposeCmd(projectDir *string) *cobra.Command {
	var (
		outPath   string
		inPath    string
		service   string
		imageExpr string
		allBuilt  bool
	)

	cmd := &cobra.Command{
		Use:   "gen-image-compose",
		Short: locale.T("genimg.short"),
		Long:  locale.T("genimg.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, root, err := loadCfg(projectDir)
			if err != nil {
				return err
			}
			composeRel := cfg.ComposeFile
			if strings.TrimSpace(inPath) != "" {
				composeRel = strings.TrimSpace(inPath)
			}
			inAbs := filepath.Join(root, filepath.FromSlash(composeRel))
			b, err := os.ReadFile(inAbs)
			if err != nil {
				return fmt.Errorf("%s: %w", locale.Tf("genimg.err.read", inAbs), err)
			}

			svc := strings.TrimSpace(service)
			if !allBuilt {
				if svc == "" {
					svc = cfg.ComposeService
				}
				if svc == "" {
					return fmt.Errorf("%s", locale.T("genimg.err.no_service"))
				}
			}

			out := strings.TrimSpace(outPath)
			if out == "" {
				out = "docker-compose.image.yml"
			}
			outAbs := filepath.Join(root, filepath.FromSlash(out))

			opts := composeimage.Options{
				TargetService: svc,
				ImageExpr:     strings.TrimSpace(imageExpr),
				AllBuilt:      allBuilt,
			}
			if allBuilt && len(cfg.DeployImages) > 0 {
				opts.ServiceImages = maps.Clone(cfg.DeployImages)
			}
			gen, err := composeimage.GenerateForArtifacts(b, opts)
			if err != nil {
				return fmt.Errorf("%s: %w", locale.T("genimg.err.transform"), err)
			}
			header := []byte(locale.T("genimg.header") + "\n")
			body := append(header, gen...)
			if err := os.WriteFile(outAbs, body, 0o644); err != nil {
				return fmt.Errorf("%s: %w", locale.Tf("genimg.err.write", outAbs), err)
			}
			fmt.Fprint(cmd.OutOrStdout(), locale.Tf("genimg.wrote", outAbs)+"\n")
			return nil
		},
	}

	cmd.Flags().StringVarP(&outPath, "output", "o", "docker-compose.image.yml", locale.T("genimg.flag.output"))
	cmd.Flags().StringVar(&inPath, "compose-file", "", locale.T("genimg.flag.compose_file"))
	cmd.Flags().StringVar(&service, "service", "", locale.T("genimg.flag.service"))
	cmd.Flags().StringVar(&imageExpr, "image-expr", "${DEPLOY_IMAGE}", locale.T("genimg.flag.image_expr"))
	cmd.Flags().BoolVar(&allBuilt, "all-built", false, locale.T("genimg.flag.all_built"))
	return cmd
}
