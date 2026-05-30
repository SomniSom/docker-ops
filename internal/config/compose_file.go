package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ArtifactsComposeFileName is the compose file delivered on artifacts deploy.
const ArtifactsComposeFileName = "docker-compose.image.yml"

// ModeIsArtifacts reports deploy_mode == artifacts (case-insensitive).
func ModeIsArtifacts(c *Config) bool {
	if c == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(c.DeployMode), "artifacts")
}

// EffectiveComposeFile returns the compose file path (relative to project root) used by
// dq ps/up/down/logs/exec so API and compose deploy target the same project definition.
func EffectiveComposeFile(c *Config, projectRoot string) string {
	if c == nil {
		return "docker-compose.yml"
	}
	cf := strings.TrimSpace(c.ComposeFile)
	if cf == "" {
		cf = "docker-compose.yml"
	}
	if !ModeIsArtifacts(c) {
		return cf
	}
	if strings.EqualFold(cf, ArtifactsComposeFileName) {
		return ArtifactsComposeFileName
	}
	if projectRoot != "" {
		p := filepath.Join(projectRoot, ArtifactsComposeFileName)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return ArtifactsComposeFileName
		}
	}
	// Artifacts on remote without local image compose: same filename as on server after deploy.
	if projectRoot == "" {
		return ArtifactsComposeFileName
	}
	return cf
}
