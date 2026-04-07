package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SomniSom/docker-ops/internal/locale"
	"gopkg.in/yaml.v3"
)

// ValidateFile parses docker-ops YAML and runs semantic checks (for `dq validate`).
func ValidateFile(absPath string) error {
	b, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("%s: %w", locale.Tf("validate.err.read_file", absPath), err)
	}
	return ValidateBytes(absPath, b)
}

// ValidateBytes parses YAML and validates. label is a path (used for messages and app_config resolution).
func ValidateBytes(label string, content []byte) error {
	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return FormatYAMLParseError(label, content, err)
	}
	if err := validateDeployMode(&cfg); err != nil {
		return err
	}
	root := filepath.Dir(label)
	if err := validateAppConfigPath(&cfg, root); err != nil {
		return err
	}
	return nil
}

func validateDeployMode(cfg *Config) error {
	dm := strings.TrimSpace(strings.ToLower(cfg.DeployMode))
	if dm != "" && dm != "source" && dm != "artifacts" {
		return fmt.Errorf("%s", locale.Tf("validate.deploy_mode", cfg.DeployMode))
	}
	return nil
}

func validateAppConfigPath(cfg *Config, projectRoot string) error {
	if strings.TrimSpace(cfg.AppConfig) == "" {
		return nil
	}
	p := cfg.AppConfig
	if !filepath.IsAbs(p) {
		p = filepath.Join(projectRoot, p)
	}
	if _, err := os.Stat(p); err != nil {
		return fmt.Errorf("%s: %w", locale.Tf("validate.app_config_missing", cfg.AppConfig, p), err)
	}
	return nil
}
