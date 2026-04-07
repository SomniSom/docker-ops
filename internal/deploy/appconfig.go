package deploy

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/sshexec"
)

// RemoteRelAppConfig is the path of the app config file relative to remote_path (for test -f on the server).
func RemoteRelAppConfig(cfg *config.Config) string {
	if cfg == nil || strings.TrimSpace(cfg.AppConfig) == "" {
		return ""
	}
	p := filepath.Clean(strings.TrimSpace(cfg.AppConfig))
	if filepath.IsAbs(p) {
		return filepath.ToSlash(filepath.Base(p))
	}
	return filepath.ToSlash(p)
}

// UploadAppConfig copies the app config from the project root to the remote tree when the file exists.
func UploadAppConfig(c *sftp.Client, projectRoot, remoteRoot string, cfg *config.Config) error {
	if cfg == nil || strings.TrimSpace(cfg.AppConfig) == "" {
		return nil
	}
	p := strings.TrimSpace(cfg.AppConfig)
	localAbs := p
	if !filepath.IsAbs(p) {
		localAbs = filepath.Join(projectRoot, p)
	}
	st, err := os.Stat(localAbs)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprint(os.Stderr, locale.Tf("deploy.appcfg.warn", localAbs))
			return nil
		}
		return err
	}
	if st.IsDir() {
		return fmt.Errorf("%s", locale.Tf("deploy.appcfg.dir", localAbs))
	}
	rel := RemoteRelAppConfig(cfg)
	rem := remoteJoin(remoteRoot, rel)
	parent := path.Dir(filepath.ToSlash(rem))
	if parent != "." && parent != "/" {
		if err := sftpMkdirAll(c, parent); err != nil {
			return err
		}
	}
	return uploadFile(c, localAbs, rem, st)
}

// RemoteConfigCheckScript returns a bash fragment: no-op if app_config unset, else test -f.
func RemoteConfigCheckScript(cfg *config.Config, remoteRoot string) string {
	rel := RemoteRelAppConfig(cfg)
	if rel == "" {
		return "true"
	}
	path := remoteJoin(remoteRoot, rel)
	return fmt.Sprintf("test -f %s", sshexec.ShellQuote(path))
}
