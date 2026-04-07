package config

import (
	"fmt"
	"strings"

	"github.com/SomniSom/docker-ops/internal/locale"
)

// RemoteYAMLTemplate generates docker-ops.yaml body (§5.4).
func RemoteYAMLTemplate(cfg *Config, anonymize bool) string {
	if cfg == nil {
		cfg = &Config{}
	}
	rs := "user@host"
	rp := fmt.Sprintf("/opt/%s", defaultProjectName(cfg.ComposeProjectName))
	if !anonymize {
		if s := strings.TrimSpace(cfg.RemoteSSH); s != "" {
			rs = s
		}
		if s := strings.TrimSpace(cfg.RemotePath); s != "" {
			rp = s
		}
	}

	pn := defaultProjectName(cfg.ComposeProjectName)
	cf := cfg.ComposeFile
	if strings.TrimSpace(cf) == "" {
		cf = "docker-compose.yml"
	}
	cs := cfg.ComposeService
	if strings.TrimSpace(cs) == "" {
		cs = "app"
	}

	var b strings.Builder
	fmt.Fprintf(&b, `%s
%s
#
`, locale.T("template.header1"), locale.T("template.header2"))
	fmt.Fprintf(&b, `
remote_ssh: %s
remote_path: %s

# Optional:
# ssh_identity: ~/.ssh/id_ed25519
# rsync_extra: --dry-run
compose_project_name: %s
compose_file: %s
compose_service: %s

# --- Artifacts deploy (deploy_mode: artifacts) ---
# deploy_mode: artifacts
# compose_file: docker-compose.image.yml
# deploy_image: ghcr.io/you/app:1.0.0
# deploy_push: true

# Extra paths relative to project root (deploy_include):
# deploy_include:
#   - assets
`, rs, rp, pn, cf, cs)
	return b.String()
}

func defaultProjectName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "app"
	}
	return s
}
