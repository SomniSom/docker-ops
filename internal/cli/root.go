// Package cli constructs the dq Cobra command tree: compose helpers, deploy, validation,
// shell completion, and man-page generation. Subcommands resolve the project root from
// docker-ops.yml (and optional --project-dir), then run docker compose either locally
// or over SSH when remote_ssh / remote_path are set in config.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/version"
)

// NewRoot returns the root *cobra.Command for the dq binary. It registers all
// subcommands (version, validate, env, compose, deploy, completion, man, etc.),
// wires persistent flags (--project-dir, --lang), and applies SilenceErrors /
// SilenceUsage so RunE failures are printed once from main without dumping full usage.
func NewRoot() *cobra.Command {
	var projectDir string

	root := &cobra.Command{
		Use:   "dq",
		Short: "Docker Quick-ops — " + locale.T("root.short"),
		Long:  fmt.Sprintf(locale.T("root.long"), version.Name, version.Version),
		// RunE errors: print once from main; do not prepend full command usage (keeps deploy/validate messages readable).
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.PersistentFlags().Changed("lang") {
				l, _ := cmd.PersistentFlags().GetString("lang")
				locale.Set(l)
			}
			return nil
		},
	}

	root.PersistentFlags().StringVar(&projectDir, "project-dir", "", locale.T("flag.project_root"))
	root.PersistentFlags().String("lang", "auto", locale.T("flag.lang"))

	root.AddCommand(newVersionCmd())
	root.AddCommand(newValidateCmd(&projectDir))
	root.AddCommand(newEnvCmd(&projectDir))
	root.AddCommand(newConfigCheckCmd(&projectDir))
	for _, c := range newComposeCmds(&projectDir) {
		root.AddCommand(c)
	}
	root.AddCommand(newDeployCmd(&projectDir))
	root.AddCommand(newGenImageComposeCmd(&projectDir))
	root.AddCommand(newCompletionCmd(root))
	root.AddCommand(newManCmd(root))

	return root
}

// loadCfg resolves the project directory from projectDir (empty means search upward
// for docker-ops.yml), loads config and returns the parsed *config.Config together
// with the absolute project root path.
func loadCfg(projectDir *string) (*config.Config, string, error) {
	root, err := config.ResolveProjectRoot(*projectDir)
	if err != nil {
		return nil, "", err
	}
	res, err := config.Load(root)
	if err != nil {
		return nil, "", err
	}
	return res.Config, res.ProjectRoot, nil
}
