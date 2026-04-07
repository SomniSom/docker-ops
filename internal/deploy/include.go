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
)

// SyncDeployIncludes uploads extra paths from the project root (readme deploy_include).
func SyncDeployIncludes(c *sftp.Client, projectRoot, remoteRoot string, cfg *config.Config) error {
	if cfg == nil || len(cfg.DeployInclude) == 0 {
		return nil
	}
	projectRoot = filepath.Clean(projectRoot)
	remoteRoot = filepath.ToSlash(filepath.Clean(remoteRoot))

	for _, raw := range cfg.DeployInclude {
		relp := strings.TrimSpace(raw)
		relp = strings.TrimPrefix(relp, "./")
		if relp == "" {
			continue
		}
		if filepath.IsAbs(relp) {
			fmt.Fprint(os.Stderr, locale.Tf("deploy.inc.skip_abs", raw))
			continue
		}
		if strings.Contains(relp, "..") {
			fmt.Fprint(os.Stderr, locale.Tf("deploy.inc.skip_unsafe", raw))
			continue
		}
		src := filepath.Join(projectRoot, filepath.FromSlash(relp))
		st, err := os.Stat(src)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprint(os.Stderr, locale.Tf("deploy.inc.missing", src))
				continue
			}
			return err
		}
		if st.IsDir() {
			if err := syncIncludeDir(c, remoteRoot, relp, src); err != nil {
				return err
			}
			continue
		}
		parent := path.Dir(filepath.ToSlash(relp))
		if parent != "." {
			remParent := remoteJoin(remoteRoot, parent)
			if err := sftpMkdirAll(c, remParent); err != nil {
				return err
			}
		}
		rem := remoteJoin(remoteRoot, relp)
		if err := uploadFile(c, src, rem, st); err != nil {
			return fmt.Errorf("%s: %w", locale.Tf("deploy.inc.err", relp), err)
		}
	}
	return nil
}

func syncIncludeDir(c *sftp.Client, remoteRoot, relp, srcDir string) error {
	return filepath.WalkDir(srcDir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		sub, err := filepath.Rel(srcDir, p)
		if err != nil {
			return err
		}
		if sub == "." {
			return nil
		}
		sub = filepath.ToSlash(sub)
		relOut := filepath.ToSlash(relp) + "/" + sub
		if d.IsDir() {
			rem := remoteJoin(remoteRoot, relOut)
			return sftpMkdirAll(c, rem)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		rem := remoteJoin(remoteRoot, relOut)
		parent := path.Dir(filepath.ToSlash(rem))
		if parent != "." && parent != "/" {
			if err := sftpMkdirAll(c, parent); err != nil {
				return err
			}
		}
		return uploadFile(c, p, rem, info)
	})
}
