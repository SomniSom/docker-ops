package config

import (
	"fmt"
	"os"

	"github.com/SomniSom/docker-ops/internal/locale"
	"gopkg.in/yaml.v3"
)

// LoadResult carries loaded config and metadata.
type LoadResult struct {
	Config     *Config
	YAMLPath   string // path to docker-ops yaml if found, else ""
	ProjectRoot string
}

// Load reads YAML (if present), merges dq.env, then process env — §14.1.
func Load(projectRoot string) (*LoadResult, error) {
	root := projectRoot
	cfg := &Config{}

	yamlPath := FindDockerOpsFile(root)
	if yamlPath != "" {
		b, err := os.ReadFile(yamlPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", locale.T("load.err.read"), err)
		}
		if err := yaml.Unmarshal(b, cfg); err != nil {
			return nil, FormatYAMLParseError(yamlPath, b, err)
		}
		if err := validateDeployMode(cfg); err != nil {
			return nil, fmt.Errorf("%s: %w", locale.Tf("load.err.config", yamlPath), err)
		}
	}

	envPath := DQEnvPath(root)
	envMap, err := ParseDQEnv(envPath)
	if err != nil {
		return nil, err
	}
	applyDQEnvMap(cfg, envMap)

	overlayProcessEnv(cfg)

	cfg.ApplyDefaults(root)

	return &LoadResult{Config: cfg, YAMLPath: yamlPath, ProjectRoot: root}, nil
}
