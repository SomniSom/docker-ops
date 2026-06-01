package cli

import (
	"github.com/SomniSom/docker-ops/internal/compose"
	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/spf13/cobra"
)

func newCleanCmd(projectDir *string) *cobra.Command {
	var allImages, volumes bool
	c := &cobra.Command{
		Use:   "clean",
		Short: locale.T("clean.short"),
		Long:  locale.T("clean.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newDockerSession(projectDir)
			if err != nil {
				return err
			}
			return runProjectClean(s, cleanOpts{AllImages: allImages, Volumes: volumes})
		},
	}
	c.Flags().BoolVarP(&allImages, "all-images", "a", false, locale.T("clean.flag.all_images"))
	c.Flags().BoolVarP(&volumes, "volumes", "v", false, locale.T("clean.flag.volumes"))
	return c
}

// newDockerSession loads config and prepares a session for plain `docker` commands
// (no Compose V2 plugin check).
func newDockerSession(projectDir *string) (*composeSession, error) {
	cfg, root, err := loadCfg(projectDir)
	if err != nil {
		return nil, err
	}
	if cfg.RemoteConfigured() {
		return &composeSession{cfg: cfg, localRoot: root}, nil
	}
	if err := compose.LookPathDocker(); err != nil {
		return nil, err
	}
	return &composeSession{cfg: cfg, localRoot: root}, nil
}
