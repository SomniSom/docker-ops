package compose

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/SomniSom/docker-ops/internal/locale"
)

// RequireComposeV2Plugin checks that `docker compose` (Compose V2 CLI plugin) is available.
// dq does not fall back to standalone `docker-compose` (V1); users should install the plugin.
func RequireComposeV2Plugin() error {
	if err := LookPathDocker(); err != nil {
		return fmt.Errorf("%s: %w\n\n%s", locale.T("compose.docker_path"), err, composeV2InstallHint())
	}
	cmd := exec.Command("docker", "compose", "version", "--short")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err == nil {
		v := strings.TrimSpace(string(out))
		if v != "" {
			return nil
		}
	}
	msg := strings.TrimSpace(stderr.String())
	if msg == "" && err != nil {
		msg = err.Error()
	}
	if looksLikeMissingComposePlugin(msg) {
		return fmt.Errorf("%s", locale.Tf("compose.v2_required", composeV2InstallHint(), msg))
	}
	if err != nil {
		return fmt.Errorf("%s: %w\nstderr: %s", locale.T("compose.version_prefix"), err, msg)
	}
	return nil
}

func looksLikeMissingComposePlugin(dockerStderr string) bool {
	s := strings.ToLower(dockerStderr)
	return strings.Contains(s, "not a docker command") ||
		strings.Contains(s, "unknown command") && strings.Contains(s, "compose") ||
		strings.Contains(s, "'compose' is not a docker command")
}

func composeV2InstallHint() string {
	return locale.T("compose.install_hint")
}

// HintIfComposePluginError appends install hint when a failed docker compose run looks like a missing plugin.
func HintIfComposePluginError(stderr string, err error) error {
	if err == nil {
		return nil
	}
	if !looksLikeMissingComposePlugin(stderr) {
		return err
	}
	return fmt.Errorf("%w\n\n%s", err, composeV2InstallHint())
}
