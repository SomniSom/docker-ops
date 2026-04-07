// Package config loads docker-ops.yaml / docker-ops.yml and dq.env per readme §14.1.
package config

import (
	"path/filepath"
	"strings"
)

// Config holds docker-ops YAML fields (snake_case in file).
type Config struct {
	ComposeProjectName string   `yaml:"compose_project_name"`
	ComposeFile        string   `yaml:"compose_file"`
	ComposeService     string   `yaml:"compose_service"`
	RemoteSSH          string   `yaml:"remote_ssh"`
	RemotePath         string   `yaml:"remote_path"`
	SSHIdentity        string   `yaml:"ssh_identity"`
	Exclude            []string `yaml:"exclude"`
	RsyncExtra         string   `yaml:"rsync_extra"`
	DeployMode         string   `yaml:"deploy_mode"`
	DeployImage        string   `yaml:"deploy_image"`
	DeployPush         *bool    `yaml:"deploy_push"`
	DeployUseRegistry  *bool    `yaml:"deploy_use_registry"`
	DeploySaveLoad     *bool    `yaml:"deploy_save_load"`
	DeploySaveCompress *bool    `yaml:"deploy_save_compress"`
	DeployInclude      []string `yaml:"deploy_include"`
	AppConfig          string   `yaml:"app_config"`
	HelpShowEffective  *bool    `yaml:"help_show_effective"`
	UseRemote          *bool    `yaml:"use_remote"` // false => force local (DOCKER_OPS_USE_REMOTE=0)
}

// RemoteConfigured reports whether remote SSH mode is available (§5.1).
func (c *Config) RemoteConfigured() bool {
	if c == nil {
		return false
	}
	if c.UseRemote != nil && !*c.UseRemote {
		return false
	}
	return strings.TrimSpace(c.RemoteSSH) != "" && strings.TrimSpace(c.RemotePath) != ""
}

// ApplyDefaults sets compose defaults from project root directory name (§5.2).
func (c *Config) ApplyDefaults(projectRoot string) {
	if c == nil {
		return
	}
	base := filepath.Base(filepath.Clean(projectRoot))
	if base == "." || base == "/" {
		base = "app"
	}
	if c.ComposeProjectName == "" {
		c.ComposeProjectName = base
	}
	if c.ComposeFile == "" {
		c.ComposeFile = "docker-compose.yml"
	}
	if c.ComposeService == "" {
		c.ComposeService = "app"
	}
}
