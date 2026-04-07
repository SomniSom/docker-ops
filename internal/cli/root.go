package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/version"
)

// NewRoot builds the dq command tree.
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
