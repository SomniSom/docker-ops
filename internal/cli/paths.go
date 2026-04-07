package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SomniSom/docker-ops/internal/locale"
)

// resolveAppConfigPath returns an absolute path for appConfig: if appConfig is already
// absolute (after Clean), it is returned as-is; otherwise it is joined under root.
func resolveAppConfigPath(root, appConfig string) string {
	appConfig = strings.TrimSpace(appConfig)
	p := filepath.Clean(appConfig)
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(root, p)
}

// checkAppConfigExists returns nil if appConfig is empty. Otherwise it requires the
// resolved path to exist and be a regular file (not a directory).
func checkAppConfigExists(root, appConfig string) error {
	if appConfig == "" {
		return nil
	}
	p := resolveAppConfigPath(root, appConfig)
	st, err := os.Stat(p)
	if err != nil {
		return fmt.Errorf("%s: %w", locale.Tf("paths.err.app_config", p), err)
	}
	if st.IsDir() {
		return fmt.Errorf("%s", locale.Tf("paths.err.app_dir", p))
	}
	return nil
}
