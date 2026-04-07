package deploy

import (
	"fmt"
	"strings"

	"github.com/pkg/sftp"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/sshexec"
)

// RunSource syncs the project tree over SFTP then runs remote reup (readme §5.3).
func RunSource(projectRoot string, cfg *config.Config) error {
	if cfg == nil || !cfg.RemoteConfigured() {
		return fmt.Errorf("%s", locale.T("err.remote_not_configured"))
	}
	rp := strings.TrimSpace(cfg.RemotePath)
	client, err := sshexec.Dial(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := sshexec.RunBash(client, "mkdir -p "+sshexec.ShellQuote(rp), false); err != nil {
		return fmt.Errorf("%s: %w", locale.T("deploy.src.remote_mkdir"), err)
	}

	c, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("%s: %w", locale.T("deploy.src.sftp"), err)
	}

	patterns := MergeExcludePatterns(cfg)
	if err := MirrorProjectTree(c, projectRoot, rp, patterns); err != nil {
		_ = c.Close()
		return err
	}
	if err := UploadAppConfig(c, projectRoot, rp, cfg); err != nil {
		_ = c.Close()
		return err
	}
	if err := SyncDeployIncludes(c, projectRoot, rp, cfg); err != nil {
		_ = c.Close()
		return err
	}
	if err := c.Close(); err != nil {
		return err
	}

	return RunRemoteReup(client, cfg)
}
